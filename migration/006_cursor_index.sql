-- Index untuk Keyset / Cursor Pagination
-- Urutan column pada index HARUS sama persis dengan ORDER BY pada query,
-- termasuk arah DESC-nya.
CREATE INDEX IF NOT EXISTS users_created_at_id_desc_idx
    ON users (created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS students_id_desc_idx
    ON students (id DESC);
