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

<!-- #### Routes
| Method | Route | Requires Login | Query Params | Post Params | Return |
|---|---|---|
| GET | / | No  | latitude=12.23&longitude=23.33 | none | List of Posts near you |
| GET  |   |   |
|   |   |   |
|   |   |   |
|   |   |   |
|   |   |   |
|   |   |   |
|   |   |   |
|   |   |   |
|   |   |   | -->
