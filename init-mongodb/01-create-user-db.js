try {
  print("Creating user...");

  // Create user
  db = db.getSiblingDB("admin");

  db.createUser({
    user: "dbuser",
    pwd: "pass123",
    roles: [{ role: "userAdminAnyDatabase", db: "admin" },
            { role: "readWrite", db: "Movies" } ],
    mechanisms: ["SCRAM-SHA-1"],
  });

  // Authenticate user
  db.auth({
    user: "dbuser",
    pwd: "pass123",
    mechanisms: ["SCRAM-SHA-1"]
  });

  // Create DB and collection
  db = new Mongo().getDB("Movies");
  db.createCollection("moviesList", { capped: false });
} catch (error) {
  print(`Failed to create DB user:\n${error}`);
}
