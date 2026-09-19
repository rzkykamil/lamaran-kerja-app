-- Tabel penyimpan sesi login. Skema wajib persis seperti ini karena
-- dipakai langsung oleh scs/pgxstore.
CREATE TABLE sessions (
    token  TEXT PRIMARY KEY,
    data   BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);
