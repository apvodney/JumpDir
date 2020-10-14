-- +migrate Up
CREATE TABLE users (
	userID bytea PRIMARY KEY,
	email varchar,
	username varchar,
	screenname varchar,
	hashAlgo integer,
	passHash varchar
);
CREATE TABLE uRegDeadline (
	secret bytea PRIMARY KEY,
	userID bytea REFERENCES users ON DELETE CASCADE,
	emailVerifyDeadline timestamptz
);
CREATE TABLE uPrefs (
	userID bytea PRIMARY KEY REFERENCES users ON DELETE CASCADE,
	avatarCircle bool
);
CREATE TABLE sessionToken (
	userID bytea PRIMARY KEY REFERENCES users ON DELETE CASCADE,
	token bytea,
	validBefore date
);
CREATE TABLE contactInfo (
	userID bytea PRIMARY KEY REFERENCES users ON DELETE CASCADE,
	id int,
	orderIndex int,
	private boolean,
	serviceName varchar,
	url varchar,
	username varchar
);
-- +migrate StatementBegin
CREATE FUNCTION attempt_start_reg(nuserID bytea, email varchar, hashAlgo integer, passHash varchar) RETURNS boolean
LANGUAGE plpgsql
AS $$
BEGIN
IF EXISTS (SELECT * FROM users WHERE userID = $1 LIMIT 1) THEN
	RETURN false;
ELSE
	INSERT INTO users (userID, email, hashAlgo, passHash)
		VALUES ($1, $2, $3, $4);
	RETURN true;
END IF;
END;
$$;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION attempt_reg_secret_insert(nsecret bytea, userID bytea, emailVerifyDeadline timestamptz) RETURNS boolean
LANGUAGE plpgsql
AS $$
BEGIN
IF EXISTS (SELECT 1 FROM uRegDeadline WHERE secret = $1) THEN
	RETURN false;
ELSE
	INSERT INTO uRegDeadline (secret, userID, emailVerifyDeadline)
		VALUES ($1, $2, $3);
	RETURN true;
END IF;
END;
$$;
-- +migrate StatementEnd

-- +migrate Down
DROP TABLE users;
DROP TABLE uRegDeadline;
DROP TABLE uPrefs;
DROP TABLE sessionToken;
DROP TABLE contactInfo;
DROP FUNCTION attempt_start_reg;
DROP FUNCTION attempt_reg_secret_insert;
