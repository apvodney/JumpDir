-- +migrate Up
CREATE TABLE users (
	userID bytea PRIMARY KEY,
	username varchar UNIQUE NOT NULL,
	email varchar NOT NULL,
	screenname varchar,
	hashAlgo integer NOT NULL,
	passHash varchar NOT NULL
);
CREATE TABLE uRegDeadline (
	secret bytea PRIMARY KEY,
	userID bytea REFERENCES users ON DELETE CASCADE,
	deadline timestamptz
);
CREATE TABLE uPrefs (
	userID bytea PRIMARY KEY REFERENCES users ON DELETE CASCADE,
	avatarCircle bool
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
CREATE FUNCTION attempt_start_reg(_username varchar, _email varchar, _hashAlgo integer, _passHash varchar, _userID bytea) RETURNS boolean
LANGUAGE plpgsql
AS $$
BEGIN
IF EXISTS (SELECT * FROM users WHERE userID = _userID LIMIT 1) THEN
	RETURN false;
ELSE
	INSERT INTO users (userID, username, email, hashAlgo, passHash)
		VALUES (_userID, _username, _email, _hashAlgo, _passHash);
	RETURN true;
END IF;
END;
$$;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION attempt_reg_secret_insert(_userID bytea, _deadline timestamptz, _secret bytea) RETURNS boolean
LANGUAGE plpgsql
AS $$
BEGIN
IF EXISTS (SELECT 1 FROM uRegDeadline WHERE secret = _secret) THEN
	RETURN false;
ELSE
	INSERT INTO uRegDeadline (secret, userID, deadline)
		VALUES (_secret, _userID, _deadline);
	RETURN true;
END IF;
END;
$$;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION finish_reg(_secret bytea, _email varchar, _now timestamptz) RETURNS bytea
LANGUAGE plpgsql
AS $$
DECLARE
	ID bytea;
BEGIN
SELECT INTO ID
	ur.userID FROM uRegDeadline ur
	INNER JOIN users u
	ON (ur.userID = u.userID)
	WHERE ur.secret = _secret AND u.email = _email AND ur.deadline > _now
	LIMIT 1;
IF ID IS NOT NULL THEN
	DELETE FROM uRegDeadline ur WHERE ur.secret = _secret;
END IF;
RETURN ID;
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
DROP TABLE uRegDeadline;
DROP TABLE uPrefs;
DROP TABLE sessionToken;
DROP TABLE contactInfo;
DROP FUNCTION attempt_start_reg;
DROP FUNCTION attempt_reg_secret_insert;
DROP FUNCTION attempt_finish_reg;
DROP FUNCTION attempt_token_insert;
