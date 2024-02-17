--! Previous: -
--! Hash: sha1:acdabbd00b2371bde1c0370c46c64d55b347997f

-- Enter migration here
CREATE TABLE users (
    username    varchar(100) PRIMARY KEY,
    password    varchar(100) NOT NULL
);
