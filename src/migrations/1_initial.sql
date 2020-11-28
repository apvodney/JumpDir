-- +migrate Up
CREATE TABLE users (
	userID bytea PRIMARY KEY,
	username varchar UNIQUE NOT NULL,
	email varchar NOT NULL,
	hashAlgo integer NOT NULL,
	passHash varchar NOT NULL
);
CREATE TABLE uRegPending (
	secret bytea,
	email varchar,
	deadline timestamptz NOT NULL,
	username varchar NOT NULL,
	hashAlgo integer NOT NULL,
	passHash varchar NOT NULL,
	PRIMARY KEY(secret, email)
);
CREATE TABLE uPrefs (
	userID bytea PRIMARY KEY REFERENCES users ON DELETE CASCADE,
	avatarCircle bool,
	screenname varchar
);
CREATE TABLE uToken (
	token bytea PRIMARY KEY,
	userID bytea REFERENCES users ON DELETE CASCADE,
	validBefore timestamptz
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
CREATE FUNCTION attempt_reg_pending_insert(_deadline timestamptz,
	_username varchar, _email varchar, _hashAlgo integer, _passHash varchar,
	_secret bytea)
RETURNS boolean
LANGUAGE plpgsql
AS $$
BEGIN
IF EXISTS (SELECT 1 FROM uRegPending WHERE secret = _secret) THEN
	RETURN false;
ELSE
	INSERT INTO uRegPending (secret, deadline, username, email, hashAlgo, passHash)
		VALUES (_secret, _deadline, _username, _email, _hashAlgo, _passHash);
	RETURN true;
END IF;
END;
$$;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION attempt_finish_reg(_secret bytea, _email varchar, _now timestamptz, _userID bytea)
RETURNS bool
LANGUAGE plpgsql
AS $$
DECLARE
	_username varchar;
	_hashAlgo integer;
	_passHash varchar;
	_deadline timestamptz;
BEGIN
IF EXISTS (SELECT * FROM users WHERE userID = _userID LIMIT 1) THEN
	RETURN false;
ELSE
	SELECT INTO STRICT _username, _hashAlgo, _passHash, _deadline
		ur.username, ur.hashAlgo, ur.passHash, ur.deadline FROM uRegPending ur
		WHERE ur.secret = _secret AND ur.email = _email;
	IF _deadline <= _now THEN
		RAISE SQLSTATE 'ZD000';
	END IF;

	INSERT INTO users (userID, username, email, hashAlgo, passHash)
		VALUES (_userID, _username, _email, _hashAlgo, _passHash);
	DELETE FROM uRegPending ur WHERE ur.secret = _secret AND ur.email = _email;
	RETURN true;
END IF;
END;
$$;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION attempt_token_insert(_userID bytea, _validBefore timestamptz, _token bytea) RETURNS boolean
LANGUAGE plpgsql
AS $$
BEGIN
IF EXISTS (SELECT 1 FROM uToken WHERE token = _token) THEN
	RETURN false;
ELSE
	INSERT INTO uToken (token, userID, validBefore)
		VALUES (_token, _userID, _validBefore);
	RETURN true;
END IF;
END;
$$;
-- +migrate StatementEnd

-- +migrate Down
DROP TABLE users;
DROP TABLE uRegPending;
DROP TABLE uPrefs;
DROP TABLE uToken;
DROP TABLE contactInfo;
DROP FUNCTION attempt_reg_pending_insert;
DROP FUNCTION attempt_finish_reg;
DROP FUNCTION attempt_token_insert;
