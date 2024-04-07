CREATE TABLE if NOT EXISTS records(
	id SERIAL PRIMARY KEY,
	identifier VARCHAR(50) UNIQUE NOT NULL,
	base VARCHAR(5) NOT NULL,
	secondary VARCHAR(5) NOT NULL,
	rate  NUMERIC  NOT NULL,
	status SMALLINT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX records_base_secondary_idx ON records(base, secondary);