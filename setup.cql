// Run the below commands to setup the cql tables.
// in production use NetworkToplogyStrategy and use a network aware snitch,
// and define data center names and replication factors for each data center
// ie CREATE KEYSPACE "Excalibur" WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'dc1' : 3, 'dc2' : 2};


CREATE KEYSPACE rent WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

USE rent;

CREATE TABLE users(
	id uuid,
	name text,
	email text,
	hash text,
	session_key text,
	PRIMARY KEY(id));

CREATE TABLE users_by_email(
	id uuid,
	name text,
	email text,
	hash text,
	session_key text,
	PRIMARY KEY(email));

CREATE MATERIALIZED VIEW users_by_session_key AS SELECT
	id, name, email, hash, session_key FROM users
	WHERE id IS NOT NULL AND session_key IS NOT NULL
	PRIMARY KEY (session_key, id);

CREATE TABLE posts(
	user_id uuid,
	post_id uuid,
	category text,
	title text,
	description text,
	price int,
	deposit int,
	available boolean,
	last_update_time timestamp,
	image_urls set<text>,
	city text,
	state text,
	latitude double,
	longitude double,
	PRIMARY KEY((state), post_id));

CREATE TABLE posts_by_category(
	user_id uuid,
	post_id uuid,
	category text,
	title text,
	description text,
	price int,
	deposit int,
	available boolean,
	last_update_time timestamp,
	image_urls set<text>,
	city text,
	state text,
	latitude double,
	longitude double,
	PRIMARY KEY((state, category), post_id));

CREATE TABLE posts_by_user(
	user_id uuid,
	post_id uuid,
	category text,
	title text,
	description text,
	price int,
	deposit int,
	available boolean,
	last_update_time timestamp,
	image_urls set<text>,
	city text,
	state text,
	latitude double,
	longitude double,
	PRIMARY KEY((user_id), post_id));

CREATE TABLE chats(
	chat_id uuid,
	post_id uuid,
	user_id uuid,
	recipient_id uuid,
	post_name text,
	main_image_url text,
	last_message_time timestamp,
	PRIMARY KEY((user_id), post_id, recipient_id));

CREATE TABLE messages_by_chat(
	chat_id uuid,
	message_time timestamp,
	message_id uuid,
	owner_id uuid STATIC,
	owner_name text STATIC,
	borrower_id uuid STATIC,
	borrower_name text STATIC,
	message_content text,
	message_sender_id uuid,
	PRIMARY KEY((chat_id), message_time, message_id)) WITH CLUSTERING ORDER BY (message_time DESC);

CREATE CUSTOM INDEX posts_index ON posts ()
USING 'com.stratio.cassandra.lucene.Index'
WITH OPTIONS = {
    'refresh_seconds' : '60',
    'schema' : '{
        fields : {
						available  					: {type : "boolean"},
            title  							: {type : "text", analyzer : "english"},
						description  				: {type : "text", analyzer : "english"},
						price		  					: {type : "integer"},
						deposit					  	: {type : "integer"},
						place 							: {type : "geo_point", latitude: "latitude", longitude: "longitude"}
        }
    }'
};

CREATE CUSTOM INDEX posts_category_index ON posts_by_category ()
USING 'com.stratio.cassandra.lucene.Index'
WITH OPTIONS = {
    'refresh_seconds' : '60',
    'schema' : '{
        fields : {
            category  					: {type : "string"},
						available  					: {type : "boolean"},
            title  							: {type : "text", analyzer : "english"},
						description  				: {type : "text", analyzer : "english"},
						price		  					: {type : "integer"},
						deposit					  	: {type : "integer"},
						place 							: {type : "geo_point", latitude: "latitude", longitude: "longitude"}
        }
    }'
};
