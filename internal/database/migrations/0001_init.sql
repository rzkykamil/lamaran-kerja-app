CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE list_items (
    id         BIGSERIAL PRIMARY KEY,
    category   TEXT NOT NULL,
    value      TEXT NOT NULL,
    sort_order INT  NOT NULL DEFAULT 0,
    UNIQUE (category, value)
);

CREATE TABLE fulltime_applications (
    id                  BIGSERIAL PRIMARY KEY,
    tgl_apply           DATE NOT NULL,
    perusahaan          TEXT NOT NULL,
    posisi              TEXT NOT NULL,
    lokasi              TEXT NOT NULL DEFAULT '',
    tipe_kerja          TEXT NOT NULL DEFAULT '',
    sumber              TEXT NOT NULL DEFAULT '',
    link_lowongan       TEXT NOT NULL DEFAULT '',
    ekspektasi_gaji     BIGINT NOT NULL DEFAULT 0,
    prioritas           TEXT NOT NULL DEFAULT 'Sedang',
    status              TEXT NOT NULL DEFAULT 'Wishlist',
    tahap_ke            INT NOT NULL DEFAULT 0,
    tgl_update_terakhir DATE NOT NULL DEFAULT CURRENT_DATE,
    next_action         TEXT NOT NULL DEFAULT '',
    tgl_follow_up       DATE,
    kontak              TEXT NOT NULL DEFAULT '',
    catatan             TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ft_status ON fulltime_applications (status);
CREATE INDEX idx_ft_update ON fulltime_applications (tgl_update_terakhir);

CREATE TABLE freelance_projects (
    id                BIGSERIAL PRIMARY KEY,
    tgl_masuk_lead    DATE NOT NULL,
    klien             TEXT NOT NULL,
    nama_project      TEXT NOT NULL,
    stack             TEXT NOT NULL DEFAULT '',
    sumber            TEXT NOT NULL DEFAULT '',
    scope_singkat     TEXT NOT NULL DEFAULT '',
    nilai_penawaran   BIGINT NOT NULL DEFAULT 0,
    status            TEXT NOT NULL DEFAULT 'Lead',
    probabilitas      NUMERIC(3,2) NOT NULL DEFAULT 0.50 CHECK (probabilitas BETWEEN 0 AND 1),
    nilai_tertimbang  BIGINT GENERATED ALWAYS AS (ROUND(nilai_penawaran * probabilitas)) STORED,
    deadline_proposal DATE,
    tgl_mulai         DATE,
    deadline_project  DATE,
    status_bayar      TEXT NOT NULL DEFAULT 'Belum Ditagih',
    sudah_dibayar     BIGINT NOT NULL DEFAULT 0,
    sisa_tagihan      BIGINT GENERATED ALWAYS AS (nilai_penawaran - sudah_dibayar) STORED,
    next_action       TEXT NOT NULL DEFAULT '',
    catatan           TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_fl_status ON freelance_projects (status);

CREATE TABLE dashboard_targets (
    metric_key   TEXT PRIMARY KEY,
    target_value INT NOT NULL DEFAULT 0
);
