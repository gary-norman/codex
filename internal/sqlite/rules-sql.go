package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

type RuleModel struct {
	DB *sql.DB
}

// CreateRule inserts a new rule into the Rules table
func (m *RuleModel) CreateRule(ctx context.Context, rule string) (int64, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction for CreateRule: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	var ruleID int64
	query := "INSERT INTO Rules (Rule, Created, Predefined) VALUES (?, DateTime('now'), 0)"
	result, err := tx.ExecContext(ctx, query, rule)
	if err != nil {
		return 0, fmt.Errorf("failed to create rule: %w", err)
	}

	// Get the last inserted ID
	ruleID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("failed to commit transaction for CreateRule: %w", err)
	}

	return ruleID, nil
}

// InsertRule inserts a rule:channel reference into the ChannelsRules table
func (m *RuleModel) InsertRule(ctx context.Context, channelID, ruleID int64) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for InsertRule: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := "INSERT INTO ChannelsRules (ChannelID, RuleID) VALUES (?, ?)"
	_, err = tx.ExecContext(ctx, query, channelID, ruleID)
	if err != nil {
		return fmt.Errorf("failed to insert rule %d for channel %d: %w", ruleID, channelID, err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction for InsertRule: %w", err)
	}

	return nil
}

// InsertChannelRule adds an existing rule to the ChannelsRules table, omitting if it already exists
func (m *RuleModel) InsertChannelRule(ctx context.Context, channelID, ruleID int64) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for InsertChannelRule: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := "INSERT INTO ChannelsRules (ChannelID, RuleID) VALUES (?, ?) ON CONFLICT(ChannelID, RuleID) DO NOTHING"
	if _, err = tx.ExecContext(ctx, query, channelID, ruleID); err != nil {
		return fmt.Errorf("failed to insert channel rule %d for channel %d: %w", ruleID, channelID, err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction for InsertChannelRule: %w", err)
	}

	return nil
}

// EditRule edits the rule string in the Rules table
func (m *RuleModel) EditRule(ctx context.Context, id int64, rule string) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for EditRule: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := "UPDATE Rules SET Rule = ? WHERE ID = ?"
	_, err = tx.ExecContext(ctx, query, rule, id)
	if err != nil {
		return fmt.Errorf("failed to edit rule %d: %w", id, err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction for EditRule: %w", err)
	}

	return nil
}

// DeleteRule removes a rule/channel reference from the ChannelsRules table
func (m *RuleModel) DeleteRule(ctx context.Context, channelID, ruleID int64) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for DeleteRule: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := "DELETE FROM ChannelsRules WHERE ChannelID = ? AND RuleID = ?"
	_, err = tx.ExecContext(ctx, query, channelID, ruleID)
	if err != nil {
		return fmt.Errorf("failed to delete channel rule %d for channel %d: %w", ruleID, channelID, err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction for DeleteRule: %w", err)
	}

	return nil
}

// All returns every row from the Rules table ordered by ID, descending
func (m *RuleModel) All(ctx context.Context) ([]models.Rule, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for All: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := "SELECT * FROM Rules ORDER BY ID DESC"
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all rules: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rule rows: %v", closeErr)
		}
	}()

	var Rules []models.Rule
	for rows.Next() {
		r := models.Rule{}
		scanErr := rows.Scan(&r.ID, &r.Rule, &r.Created, &r.Predefined)
		if scanErr != nil {
			return nil, scanErr
		}
		Rules = append(Rules, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rule rows: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction for DeleteRule: %w", err)
	}

	return Rules, nil
}

func (m *RuleModel) AllForChannel(ctx context.Context, channelID int64) ([]models.Rule, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for AllForChannel in Comments: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarn("Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	// fetch the references from ChannelsRules
	query := "SELECT RuleID FROM ChannelsRules WHERE ChannelID = ?"
	rows, err := tx.QueryContext(ctx, query, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query rules for channel %d: %w", channelID, err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			models.LogWarn("Failed to close rows: %v", closeErr)
		}
	}()
	var IDs []int
	for rows.Next() {
		var i int
		err := rows.Scan(&i)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rule ID: %w", err)
		}
		IDs = append(IDs, i)
	}

	// prepare the statement for use in the loop
	rulequery, insertErr := m.DB.Prepare("SELECT * FROM Rules WHERE ID = ?")
	if insertErr != nil {
		return nil, insertErr
	}
	defer func(query *sql.Stmt) {
		closErr := query.Close()
		if closErr != nil {
			models.LogWarn("Failed to close prepared statement in AllForChannel: %v", closErr)
		}
	}(rulequery)

	var rules []models.Rule
	// create a []rule from the slice of rule IDs
	for _, ruleID := range IDs {
		rows, err := rulequery.QueryContext(ctx, ruleID)
		if err != nil {
			return rules, fmt.Errorf("failed to query rule %d: %w", ruleID, err)
		}
		for rows.Next() {
			r := models.Rule{}
			scanErr := rows.Scan(&r.ID, &r.Rule, &r.Created, &r.Predefined)
			if scanErr != nil {
				return nil, fmt.Errorf("failed to scan rule %d: %w", ruleID, scanErr)
			}
			rules = append(rules, r)
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction for AllForChannel in Comments: %w", err)
	}

	return rules, nil
}
