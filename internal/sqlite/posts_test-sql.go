package sqlite

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/gary-norman/forum/internal/models"
)

func TestPostModelAll(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		DB *sql.DB
	}
	tests := []struct {
		name    string
		fields  fields
		want    []models.Post
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &PostModel{
				DB: tt.fields.DB,
			}
			got, err := m.All(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("All() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("All() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostModelGetPostsByChannel(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		channel int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.Post
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &PostModel{
				DB: tt.fields.DB,
			}
			got, err := m.GetPostsByChannel(ctx, tt.args.channel)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPostsByChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPostsByChannel() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostModelInsert(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		title        string
		content      string
		images       string
		author       string
		authorAvatar string
		authorID     models.UUIDField
		commentable  bool
		isFlagged    bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &PostModel{
				DB: tt.fields.DB,
			}
			_, err := m.Insert(
				ctx,
				tt.args.title,
				tt.args.content,
				tt.args.images,
				tt.args.author,
				tt.args.authorAvatar,
				tt.args.authorID,
				tt.args.commentable,
				tt.args.isFlagged,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
