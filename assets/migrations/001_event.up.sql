CREATE TABLE IF NOT EXISTS events
(
    client_time DATETIME,
    server_time DATETIME,
    ip          IPv4,
    device_id   String,
    device_os   String,
    session     String,
    sequence    Int16,
    event_type  String,
    param_int   Int32,
    param_str   String
) Engine = MergeTree
      ORDER BY server_time;