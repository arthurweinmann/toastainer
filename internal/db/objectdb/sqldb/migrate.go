package sqldb

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/toastate/toastainer/internal/utils"

	_ "embed"
)

//go:embed migrations/*
var migratefolder embed.FS

const schema = `CREATE TABLE IF NOT EXISTS toastainermeta(
	id VARCHAR(32) NOT NULL,
	data VARCHAR(1024) NOT NULL default '',
	PRIMARY KEY(id)
)`

// MigrateTimeFormat defines the timeformat as required in c.db.migrate scripts
const MigrateTimeFormat = "2006-01-02-15-04"

func (c *Client) initMigrate() error {
	_, err := c.db.Exec(schema)
	if err != nil {
		return err
	}

	var lock []int
	err = c.db.Select(&lock, "SELECT GET_LOCK('migrate', 5);")
	if err != nil {
		return err
	}

	if lock[0] == 0 {
		return fmt.Errorf("Could not acquire lock")
	}
	defer c.db.Exec("SELECT RELEASE_LOCK('migrate')")

	var lastMigrate []string
	lastMigrateDate := time.Time{}
	err = c.db.Select(&lastMigrate, "SELECT data FROM toastainermeta WHERE id=?", "lastMigrate")
	if err != nil {
		return err
	}

	if len(lastMigrate) == 0 {
		utils.Info("msg", "no last migration date found")
	} else {
		var err error
		lastMigrateDate, err = time.Parse(MigrateTimeFormat, lastMigrate[0])
		if err != nil {
			return err
		}
		utils.Info("msg", "last migration date", "date", lastMigrateDate.Format(MigrateTimeFormat))
	}

	newTime, err := c.performMigration(lastMigrateDate)
	if err != nil {
		return err
	}

	if newTime.Equal(time.Time{}) {
		return nil
	}

	_, err = c.db.Exec("REPLACE INTO toastainermeta(id, data) VALUES (?,?)", "lastMigrate", newTime.Format(MigrateTimeFormat))
	return err
}

func (c *Client) performMigration(from time.Time) (time.Time, error) {
	direntries, _ := migratefolder.ReadDir("migrations")

	var dir []fs.FileInfo

	for i := 0; i < len(direntries); i++ {
		info, err := direntries[i].Info()
		if err != nil {
			return time.Time{}, err
		}
		dir = append(dir, info)
	}

	if len(dir) == 0 {
		utils.Info("msg", "no sql migration found")
		return time.Time{}, nil
	}

	// Filter
	{
		s2 := dir
		dir = dir[:0]
		for _, v := range s2 {
			if v.IsDir() {
				continue
			}
			if v.Name()[0] == '.' {
				continue
			}
			dir = append(dir, v)
		}
	}

	if len(dir) == 0 {
		utils.Info("msg", "no sql migration found")
		return time.Time{}, nil
	}

	sort.Sort(alphabetical(dir))

	// Skip already done
	{
		skip := 0
		for k, file := range dir {
			date, err := migrationDate(file.Name())
			if err != nil {
				return time.Time{}, err
			}
			if date.Equal(from) || date.Before(from) {
				skip = k + 1
				continue
			}
			break
		}
		dir = dir[skip:]
		if len(dir) == 0 {
			utils.Info("msg", "no sql migration required")
			return time.Time{}, nil
		}
	}

	for _, file := range dir {
		fname := file.Name()
		var execs []string
		{
			fcont, err := migratefolder.ReadFile(filepath.Join("migrations", fname))
			if err != nil {
				return time.Time{}, err
			}
			execs = strings.Split(string(fcont), "\n\n\n\n")
		}
		execLen := len(execs)
		for k, v := range execs {
			utils.Info("msg", "migration", "file", fname, "status", fmt.Sprintf("%d/%d", k+1, execLen))
			_, err := c.db.Exec(v)
			if err != nil {
				return time.Time{}, err
			}
		}
	}
	return migrationDate(dir[len(dir)-1].Name())
}

func migrationDate(in string) (time.Time, error) {
	split := strings.Split(in, "_")
	out, err := time.Parse(MigrateTimeFormat, split[0])
	if err != nil {
		return time.Time{}, err
	}
	return out, nil
}

type alphabetical []os.FileInfo

func (ap alphabetical) Len() int {
	return len(ap)
}

func (ap alphabetical) Less(i, j int) bool {
	return ap[i].Name() < ap[j].Name()
}

func (ap alphabetical) Swap(i, j int) {
	ap[i], ap[j] = ap[j], ap[i]
}
