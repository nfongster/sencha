CREATE TABLE phases (
    number INTEGER PRIMARY KEY,
    name   TEXT NOT NULL
);

CREATE TABLE levels (
    number        INTEGER PRIMARY KEY,
    phase_number  INTEGER NOT NULL REFERENCES phases(number),
    grammar_md    TEXT NOT NULL,
    exceptions_md TEXT
);

CREATE TABLE vocabulary (
    id            SERIAL PRIMARY KEY,
    level_number  INTEGER NOT NULL REFERENCES levels(number),
    korean        TEXT NOT NULL,
    english       TEXT NOT NULL
);

CREATE TABLE sentences (
    id            SERIAL PRIMARY KEY,
    level_number  INTEGER NOT NULL REFERENCES levels(number),
    korean        TEXT NOT NULL,
    english       TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
