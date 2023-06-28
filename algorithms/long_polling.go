// Long polling in golang
import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Endpoint for long polling
	r.GET("/longpoll", func(c *gin.Context) {
		// Create a channel to track client disconnection
		disconnect := make(chan struct{})
		// Create a channel to send response to the client
		response := make(chan string)

		// Start a goroutine to handle client disconnection
		go func() {
			// Wait until the client disconnects
			<-c.Writer.CloseNotify()
			// Signal the disconnect channel
			close(disconnect)
		}()

		// Start a goroutine to simulate some processing
		go func() {
			// Simulate some processing
			// Replace this with your actual logic
			// For example, you can fetch data from a database or perform some computations
			// Here, we sleep for 5 seconds to simulate processing time
			time.Sleep(5 * time.Second)

			// Check if the client has disconnected
			select {
			case <-disconnect:
				// Client has disconnected, do not send the response
				return
			default:
				// Client is still connected, send the response
				response <- "Data to send to the client"
			}
		}()

		// Wait for either the response or client disconnection
		select {
		case data := <-response:
			// Send the response to the client
			c.String(http.StatusOK, data)
		case <-disconnect:
			// Client has disconnected
			// You can perform any cleanup or logging here
			fmt.Println("Client disconnected")
		}
	})

	// Run the server
	r.Run(":8080")
}