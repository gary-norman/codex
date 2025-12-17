package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/gary-norman/forum/internal/models"
)

func TestReactionModelUpsert(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		liked            bool
		disliked         bool
		authorID         models.UUIDField
		reactedPostID    int64
		reactedCommentID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Valid upsert post reaction",
			fields: fields{
				DB: setupTestDB(t), // Your test DB or mock setup
			},
			args: args{
				liked:            true,
				disliked:         false,
				authorID:         models.NewUUIDField(), // use your UUID generator
				reactedPostID:    1,
				reactedCommentID: 0,
			},
			wantErr: false,
		},
		{
			name: "Invalid both post and comment",
			fields: fields{
				DB: setupTestDB(t),
			},
			args: args{
				liked:            false,
				disliked:         true,
				authorID:         models.NewUUIDField(),
				reactedPostID:    1,
				reactedCommentID: 2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ReactionModel{
				DB: tt.fields.DB,
			}
			err := m.Upsert(ctx, tt.args.liked, tt.args.disliked, tt.args.authorID, tt.args.reactedPostID, tt.args.reactedCommentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
