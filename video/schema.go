package video

const dbFile = "video.db"

var InitialMigration = `
-- +migrate Up

-- +migrate StatementBegin
CREATE TABLE IF NOT EXISTS video (
    "sd_hash" TEXT PRIMARY KEY,

    "created_at" TEXT NOT NULL,

    "url" TEXT NOT NULL,
    "path" TEXT NOT NULL,
    "remote_path" TEXT NOT NULL DEFAULT "",

	"type" TEXT NOT NULL,
	"channel" TEXT NOT NULL,

	"last_accessed" TIMESTAMP,
	"access_count" INTEGER NOT NULL DEFAULT 0,

	"size" INTEGER NOT NULL DEFAULT 0,
	"checksum" TEXT
);
-- +migrate StatementEnd

-- +migrate Down

-- +migrate StatementBegin
DROP TABLE video;
-- +migrate StatementEnd
`
