CREATE TABLE ops_space (
  id INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
  name VARCHAR(200) NOT NULL,
  owner VARCHAR(30) DEFAULT '',
  acl VARCHAR(30) DEFAULT 'open',
  "DESC" CLOB,
  createdBy VARCHAR(30) DEFAULT '',
  createdDate DATETIME,
  editedBy VARCHAR(30) DEFAULT '',
  editedDate DATETIME,
  deleted INT DEFAULT 0,
  CONSTRAINT uk_name UNIQUE (name)
);

