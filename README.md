# sencha
An AI-powered language-learning app.

## PostgreSQL Setup (Phase 3)

```bash
# Install server and client tools
sudo apt update && sudo apt install -y postgresql postgresql-client

# Start the service
sudo systemctl enable postgresql --now

# Create the database
sudo -u postgres createdb sencha

# Set a password for the postgres user
sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'sencha';"
```

Add `database_url` to `backend/config.json`:

```json
{
  "llm": {
    "base_url": "http://localhost:11434/v1",
    "model": "llama3.2"
  },
  "database_url": "postgres://postgres:sencha@localhost:5432/sencha?sslmode=disable"
}
```

When `database_url` is set, the server connects to PostgreSQL, runs pending migrations, and uses the database for storage. When omitted, the server uses an in-memory repository with auto-seeded defaults (backward-compatible).
