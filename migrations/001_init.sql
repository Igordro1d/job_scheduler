CREATE TABLE jobs (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    type            TEXT         NOT NULL,
    payload         JSONB        NOT NULL DEFAULT '{}'::jsonb,
    priority        INT          NOT NULL DEFAULT 0,
    status          TEXT         NOT NULL DEFAULT 'pending'
                                 CHECK (status IN (
                                     'pending',
                                     'in_progress',
                                     'completed',
                                     'failed',
                                     'dead'
                                 )),
    depends_on      UUID         REFERENCES jobs(id),
    max_retries     INT          NOT NULL DEFAULT 3,
    retry_count     INT          NOT NULL DEFAULT 0,
    idempotency_key TEXT         UNIQUE,
    locked_by       TEXT,
    locked_at       TIMESTAMPTZ,
    run_after       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_jobs_claim
    ON jobs (priority DESC, created_at)
    WHERE status = 'pending';
