## Rent Server

A backend for a product rental service. Includes:
- a Users model
- a Posts model
- a Chats model

#### Features
- Users can have 1 on 1 chats.
- Users can create posts to rent out equipment.
- Users can search posts in their area.
- Authentication with sessions.

#### Tech
- Written in Go
- Cassandra backend
- Lucene for geospatial search

#### Development
1. Clone this github repo.
2. [Download & Install Cassandra](http://cassandra.apache.org/download/)
3. [Install Cassandra Lucene Plugin from Stratio](https://github.com/Stratio/cassandra-lucene-index)
4. cd to Cassandra directory and start with:
  - `bin/cassandra -f`
5. in a new terminal cd to Cassandra directory and start cqlsh with:
  - `bin/cqlsh`
6. In the cqlsh command line type
  - `source /path/to/this/cloned/directory/setup.cql`
7. The database will now be initialized with the correct keyspaces and tables. Next cd to this cloned directory and run the command
  - `go install && rent`

You can now hit the local server at `localhost:8080`

#### Routes
| Method | Route | Requires Login | Query Params | Post Params | Return |
|---|---|---|---|---|---|
| GET | /posts | No  | latitude=?&longitude=? | none | List of Posts nearby |
| GET  | /posts/category/:category  | no | latitude=?&longitude=? | none | List of Posts nearby by Category |
| GET  | /posts/user/:user_id  | no  | none | none | List of posts for a user |
| GET  |  /posts/state/:state/:post_id | no  | none | none | a post |
| POST  | /posts/create or /posts/edit | yes  | none | post_id, category, title, description, price, deposit, available, city, state, latitude, longitude, user_id | Success or Error message on an edit or create |
| DELETE   | /posts/state/:state/category/:category/user/:user_id/:post_id  | yes  | none | none | Success or Error message |
| POST  | /register  | no  | none | name, password and email | Success or Error Message |
| POST  | /login  | no  | none | email, password | Redirect to Posts Index
| POST  | /logout  | yes  | none | email | Redirect to Posts Index
| POST  | /user/edit  | yes  | none | email, name | Success or Error Message |
| DELETE | /user/:user_id | yes | none | none | Success or Error Message |
| GET | /chats/user/:user_id | yes | none | none | List of chats for the user |
| GET | /chats/user/:user_id/:chat_id | yes | none | none | List of messages for a chat |
| POST | /chats/create | yes | none | user_id, recipient_id, post_id, post_name, main_image_url, recipient_name, message_content | Success or Error message |
| POST | /chats/message/create | yes | none | user_id, chat_id, message_content, post_id, recipientId | Success or Error Message |
