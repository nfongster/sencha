
#### Phase 1: Setup the Console App and Go Server

1. **Initialize Project Structure**
   - Create a new directory for your project (e.g., `language-learning-app`).
   - Inside this directory, create subdirectories for the console app (`console`) and the Go server (`server`).

2. **Console App Implementation**
   - Write a simple console application in Go.
   - The console app should have two commands: `Start` and `Quit`.

3. **Go Server Implementation**
   - Create a Go server that listens on a port (e.g., 8080).
   - The server will handle incoming requests from the console app.

#### Phase 2: Implement Study Session Logic

1. **Hard-Coded Data**
   - Define a list of 10 example sentences in JSON format.
   - Store this data in a Go file within the `server` directory.

2. **Randomization**
   - Write logic in the server to randomly select sentences from the hard-coded list.

3. **Console App Interaction**
   - In the console app, upon selecting `Start`, initiate a study session by connecting to the Go server.
   - The console app should display a Korean or English sentence.
   - Wait for user input (spacebar press to reveal answer).

4. **User Response Handling**
   - After revealing the answer, prompt the user to select one of three options: `Pass`, `Hard`, or `Fail`.
   - Based on the user's selection, send the corresponding response back to the server.

#### Phase 3: Server Logic

1. **Receive Study Session Requests**
   - The Go server should receive a request to start a study session.
   - Randomly select a sentence from the hard-coded list and return it to the console app.

2. **Handle User Responses**
   - Receive the user's response (`Pass`, `Hard`, or `Fail`).
   - For now, ignore this response as the system will not track or time the user's responses.

3. **Session Completion**
   - After all 10 sentences have been shown, the server can close the connection and return to listening for new requests.

#### Phase 4: Testing

1. **Run the Console App**
   - Compile and run the console app.
   - Test each command (`Start` and `Quit`) to ensure they work as expected.

2. **Test the Study Session**
   - Start a study session by selecting `Start`.
   - Verify that sentences are displayed, and responses are handled correctly.

#### Phase 5: Final Touches

1. **Error Handling**
   - Add basic error handling for network communication between the console app and server.

2. **Documentation**
   - Document the project structure, setup instructions, and how to run the application.
