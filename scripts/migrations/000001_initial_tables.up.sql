BEGIN;

CREATE TABLE commands(
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source TEXT NOT NULL,

    -- 'running' | 'error' | 'finished'
    status TEXT NOT NULL,

    -- useful for describing error
    status_desc TEXT DEFAULT '',

    -- output of the script
    output TEXT DEFAULT '',

    exit_code INTEGER,
    signal INTEGER
);

COMMIT;
