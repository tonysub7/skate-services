package disk

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"skatechain.org/lib/db"
	config "skatechain.org/relayer/db"

	"skatechain.org/lib/logging"
)

var (
	TaskLogger = db.NewFileLogger(config.DbDir, "skateapp_tasks.log")
	SkateAppDB *sql.DB
)

const (
	SignedTaskSchema = "SignedTasks"
)

func init() {
	logger := logging.NewLoggerWithConsoleWriter()
	db, err := sql.Open("sqlite3", filepath.Join(config.DbDir, "skateapp.db"))
	if err != nil {
		logger.Fatal("Relayer DB initialization failed")
		panic("Relayer DB initialization failed")
	}
	SkateAppDB = db
}

type SignedTask struct {
	TaskId    uint32
	Message   string
	Initiator string
	ChainType uint32
	ChainId   uint32
	Hash      string
	Operator  string
	Signature string
}

func InitializeSkateApp() {
	SkateAppDB.Exec(`CREATE TABLE IF NOT EXISTS ? (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
	  taskId       TEXT,
	  message      TEXT,
	  initiator    TEXT,
	  chainId      INTEGER,
	  chainType    TEXT,
	  hash         TEXT,
	  operator     TEXT,
	  signature    TEXT
	)`, SignedTaskSchema)
}

func InsertSignedTask(signedTask SignedTask) error {
	_, err := SkateAppDB.Exec(
		"INSERT INTO ? (taskId, message, initiator, chainId, chainType, hash, operator, signature) VALUES (?,?,?,?,?,?,?,?)",
		SignedTaskSchema,
		signedTask.TaskId, signedTask.Message, signedTask.Initiator, signedTask.ChainId,
		signedTask.ChainType, signedTask.Hash, signedTask.Operator, signedTask.Signature,
	)
	if err != nil {
		TaskLogger.Error("InsertTask failed", "task", signedTask, "err", err)
		return err
	}
	return nil
}

func RetrieveSignedTasks() ([]SignedTask, error) {
	rows, err := SkateAppDB.Query("SELECT * FROM ?", SignedTaskSchema)
	if err != nil {
		TaskLogger.Error("SelectAllTasks failed", "err", err)
		return nil, err
	}
	defer rows.Close()

	var resultTasks []SignedTask

	for rows.Next() {
		var task SignedTask
		var entryid int

		err := rows.Scan(
			&entryid, &task.TaskId, &task.Message, &task.Initiator,
			&task.ChainId, &task.ChainType, &task.Hash, &task.Operator, &task.Signature,
		)
		if err != nil {
			return nil, err
		}
		TaskLogger.Info("Task", "task", task)
		resultTasks = append(resultTasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resultTasks, nil
}
