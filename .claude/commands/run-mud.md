# Run Matrix MUD Development Server

Start the Matrix MUD development server with proper output and connection information.

## Steps

1. Build the latest version:
   ```bash
   make build
   ```

2. If build succeeds, start the server:
   ```bash
   ./bin/matrix-mud
   ```

3. Display connection information to the user:
   - Telnet server: `localhost:2323`
   - Web interface: `http://localhost:8080`
   - Admin console: `localhost:9090`

4. Inform the user how to connect:
   ```bash
   telnet localhost 2323
   ```

5. Mention that the server will run in the foreground and can be stopped with Ctrl+C

## Example Output

```
Building Matrix MUD...
Build successful!

Starting Matrix MUD server...
Matrix Construct Server v1.28 (Phase 28) started on port 2323...

Server is now running:
  Telnet:  localhost:2323 (connect with: telnet localhost 2323)
  Web:     http://localhost:8080
  Admin:   localhost:9090

Press Ctrl+C to stop the server.
```
