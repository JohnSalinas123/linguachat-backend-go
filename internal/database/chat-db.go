package database

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
)

func (pg *postgres) PostNewChatFromInvite(ctx context.Context, userID string, inviteCode string) (bool, error) {

	// retrieve userID of invite creator
	inviteCreatorQuery := `SELECT creator_id, exp_date, consumed
		FROM invite
		WHERE invite_code=$1`

	var inviteDetails struct {
		creator_id 	string
		exp_date 	time.Time
		consumed	bool
	}


	err := pg.db.QueryRow(ctx, inviteCreatorQuery, inviteCode).Scan(&inviteDetails.creator_id, &inviteDetails.exp_date, &inviteDetails.consumed )
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, fmt.Errorf("invite not found: %w", err)
		}
		return false, fmt.Errorf("unable to scan invite row: %w", err)
	}

	// validate if invite is already consumed
	if inviteDetails.consumed {
		return false ,fmt.Errorf("invite with invite_code %s already consumed", inviteCode)
	}

	// validate exp_date of invite
	if inviteDetails.exp_date.Before(time.Now()) {
		return false, fmt.Errorf("invite with invite_code %s has expired", inviteCode)
	}

	// create chat row
	createChatQuery := `INSERT INTO chat (id, created_at)
		VALUES ($1::UUID, $2)`

	chatUUID, err := uuid.NewV4()
	if err != nil {
		return false, fmt.Errorf("failed to generate invite UUID: %w", err)
	}

	_, err = pg.db.Exec(ctx, createChatQuery, chatUUID.String(), time.Now().UTC())
	if err != nil {
		return false, fmt.Errorf("failed to create new chat: %w", err)
	}

	// create chat_participant row for creator
	// role: admin
	createChatParticipantQuery := `INSERT INTO chat_participant (created_at, role, chat_id, user_id)
		VALUES ($1, $2, $3, $4)`

	_, err = pg.db.Exec(ctx, createChatParticipantQuery, time.Now().UTC(), "admin", chatUUID.String(), inviteDetails.creator_id)
	if err != nil {
		return false, fmt.Errorf("failed to create new chat_participant for creator: %w", err)
	}

	// create chat_partipant row for member
	// role: member
	_, err = pg.db.Exec(ctx, createChatParticipantQuery, time.Now().UTC(), "member", chatUUID.String(), userID)
	if err != nil {
		return false, fmt.Errorf("failed to create new chat_participant for member: %w", err)
	}

	// modify invite row of invite
	updateInviteQuery := `UPDATE invite SET chat_id=$1, consumed=$2, consumed_at=$3
		WHERE invite_code=$4`

	_, err = pg.db.Exec(ctx, updateInviteQuery,chatUUID.String(), true, time.Now().UTC(), inviteCode)
	if err != nil {
		return false, fmt.Errorf("failed to update invite row with invite code %s: %w", inviteCode, err)
	}

	return true, nil

}