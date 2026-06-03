FROM scratch
COPY pork-history-updater /pork-history-updater
ENTRYPOINT ["/pork-history-updater"]
