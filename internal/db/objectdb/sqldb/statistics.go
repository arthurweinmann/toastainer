package sqldb

import (
	"database/sql"

	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/model"
)

func (c *Client) UpsertUserStatistics(stat *model.UserStatistics) error {
	_, err := c.db.Exec("INSERT INTO user_statistics(user_id, month_year, duration_ms, cpus, executions, ram_gbs, net_ingress, net_egress) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE duration_ms = ?, cpus = ?, executions = ?, ram_gbs = ?, net_ingress = ?, net_egress = ?", stat.UserID, stat.Monthyear, stat.DurationMS, stat.CPUS, stat.Executions, stat.RAMGBS, stat.NetIngress, stat.NetEgress, stat.DurationMS, stat.CPUS, stat.Executions, stat.RAMGBS, stat.NetIngress, stat.NetEgress)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) IncrUserStatistics(stat *model.UserStatistics) error {
	_, err := c.db.Exec("INSERT INTO user_statistics(user_id, month_year, duration_ms, cpus, executions, ram_gbs, net_ingress, net_egress) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE duration_ms = duration_ms + ?, cpus = cpus + ?, executions = executions + ?, ram_gbs = ram_gbs + ?, net_ingress = net_ingress + ?, net_egress = net_egress + ?", stat.UserID, stat.Monthyear, stat.DurationMS, stat.CPUS, stat.Executions, stat.RAMGBS, stat.NetIngress, stat.NetEgress, stat.DurationMS, stat.CPUS, stat.Executions, stat.RAMGBS, stat.NetIngress, stat.NetEgress)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetUserStatistics(userid, monthyear string) (*model.UserStatistics, error) {
	stats := []model.UserStatistics{}
	err := c.db.Select(&stats, "SELECT * FROM user_statistics WHERE user_id=? AND month_year=?", userid, monthyear)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, objectdberror.ErrNotFound
		}
		return nil, err
	}
	if len(stats) == 0 {
		return nil, objectdberror.ErrNotFound
	}

	return &stats[0], nil
}

func (c *Client) GetAllUserStatistics(userid string) ([]*model.UserStatistics, error) {
	stats := []*model.UserStatistics{}
	err := c.db.Select(&stats, "SELECT * FROM user_statistics WHERE user_id=?", userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, objectdberror.ErrNotFound
		}
		return nil, err
	}
	if len(stats) == 0 {
		return nil, objectdberror.ErrNotFound
	}

	return stats, nil
}
