# Dockerfile for QuestDB
FROM questdb/questdb:6.0.4

# Expose QuestDB ports
EXPOSE 9000
EXPOSE 9009

# Start QuestDB in detached mode
CMD ["-d", "-p", "9000:9000", "-p", "9009:9009"]