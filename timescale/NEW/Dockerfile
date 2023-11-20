FROM timescale/timescaledb:latest-pg12

# Set the environment variables
ENV POSTGRES_PASSWORD=password

# Expose the PostgreSQL port
EXPOSE 5432

# Start TimescaleDB in detached mode
CMD ["-d", "--name", "timescale", "-p", "5432:5432"]