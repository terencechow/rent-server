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
	name text,
	description text,
	price int,
	deposit int,
	minimum_rental_days int,
	next_available_date timestamp,
	image_urls set<text>,
	city text,
	state text,
	latitude float,
	longitude float,
	PRIMARY KEY((state, post_id), next_available_date)) WITH CLUSTERING ORDER BY (next_available_date DESC);

CREATE TABLE posts_by_category(
	user_id uuid,
	post_id uuid,
	category text,
	name text,
	description text,
	price int,
	deposit int,
	minimum_rental_days int,
	next_available_date timestamp,
	image_urls set<text>,
	city text,
	state text,
	latitude float,
	longitude float,
	PRIMARY KEY((state, category), next_available_date, post_id)) WITH CLUSTERING ORDER BY (next_available_date DESC);

CREATE TABLE posts_by_user(
	user_id uuid,
	post_id uuid,
	category text,
	name text,
	description text,
	price int,
	deposit int,
	minimum_rental_days int,
	next_available_date timestamp,
	image_urls set<text>,
	city text,
	state text,
	latitude float,
	longitude float,
	PRIMARY KEY((user_id), next_available_date, post_id)) WITH CLUSTERING ORDER BY (next_available_date DESC);

CREATE TABLE chats(
	chat_id uuid,
	post_id uuid,
	user_id uuid,
	post_name text,
	main_image_url text,
	last_message_time timestamp,
	PRIMARY KEY((chat_id, user_id), last_message_time)) WITH CLUSTERING ORDER BY (last_message_time DESC);

CREATE TABLE messages_by_chat(
	chat_id uuid,
	message_time timestamp,
	message_id uuid,
	post_id uuid STATIC,
	post_name text STATIC,
	owner_id uuid STATIC,
	owner_name text STATIC,
	borrower_id uuid STATIC,
	borrower_name text STATIC,
	main_image_url text STATIC,
	message_content text,
	message_sender_id uuid,
	PRIMARY KEY((chat_id), message_time, message_id)) WITH CLUSTERING ORDER BY (message_time DESC);

CREATE CUSTOM INDEX posts_index ON posts ()
USING 'com.stratio.cassandra.lucene.Index'
WITH OPTIONS = {
    'refresh_seconds' : '60',
    'schema' : '{
        fields : {
            name  							: {type : "text", analyzer : "english"},
						description  				: {type : "text", analyzer : "english"},
						price		  					: {type : "integer"},
						deposit					  	: {type : "integer"},
						minimum_rental_days : {type : "integer"},
            next_available_date : {type : "date", pattern : "yyyy/MM/dd"},
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
            name  							: {type : "text", analyzer : "english"},
						description  				: {type : "text", analyzer : "english"},
						price		  					: {type : "integer"},
						deposit					  	: {type : "integer"},
						minimum_rental_days : {type : "integer"},
            next_available_date : {type : "date", pattern : "yyyy/MM/dd"},
						place 							: {type : "geo_point", latitude: "latitude", longitude: "longitude"}
        }
    }'
};
