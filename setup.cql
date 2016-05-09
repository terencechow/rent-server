// Run the below commands to setup the cql tables.
// in production use NetworkToplogyStrategy and use a network aware snitch,
// and define data center names and replication factors for each data center
// ie CREATE KEYSPACE "Excalibur" WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'dc1' : 3, 'dc2' : 2};


CREATE KEYSPACE rent WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

USE rent;

CREATE TABLE users_by_id(
	id uuid,
	name text,
	email text,
	pass text,
	latlng varchar,
	post_ids list<timeuuid>,
	session_key text,
	session_expiry timestamp,
	chats set<uuid>,
	PRIMARY KEY(id));

CREATE TABLE posts_by_category(
	user_id uuid,
	post_id timeuuid,
	title text,
	description text,
	price int,
	available boolean,
	category text,
	images list<text>,
	latlng varchar,
	PRIMARY KEY((category), post_id)) WITH CLUSTERING ORDER BY (post_id DESC);

CREATE TABLE chats_by_id(
	chat_id uuid,
	post_id timeuuid,
	post_title text,
	post_main_image text,
	sender_name text,
	recipient_name text,
	last_update timeuuid,
	PRIMARY KEY((chat_id),last_update)) WITH CLUSTERING ORDER BY (last_update DESC);

CREATE TABLE messages_by_chat_id(
	chat_id uuid,
	message_id timeuuid,
	sender_name text,
	message_content text,
	PRIMARY KEY((chat_id), message_id)) WITH CLUSTERING ORDER BY (message_id DESC);
