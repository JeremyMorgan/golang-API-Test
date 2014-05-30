package main
 
import (
    "errors"
   //"fmt"
    "github.com/stretchr/goweb"
    "github.com/stretchr/goweb/context"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "time"
)
 
const (
    Address string = ":9090"
)
 
// mapRoutes contains lots of examples of how to map Books in
// Goweb.  It is in its own function so that test code can call it
// without having to run main().
func mapRoutes() {
 
    /*
        Add a pre-handler to save the referrer
    */
    goweb.MapBefore(func(c context.Context) error {
 
        // add a custom header
        c.HttpResponseWriter().Header().Set("X-Custom-Header", "Goweb")
 
        return nil
    })
 
    /*
        Add a post-handler to log someBook
    */
    goweb.MapAfter(func(c context.Context) error {
        // TODO: log this
        return nil
    })
 
    /*
        Map the homepage...
    */
    goweb.Map("/", func(c context.Context) error {
        return goweb.Respond.With(c, 200, []byte("Welcome to the Goweb example app - see the terminal for instructions."))
    })
 
    /*
        /status-code/xxx
        Where xxx is any HTTP status code.
    */
    goweb.Map("/status-code/{code}", func(c context.Context) error {
 
        // get the path value as an integer
        statusCodeInt, statusCodeIntErr := strconv.Atoi(c.PathValue("code"))
        if statusCodeIntErr != nil {
            return goweb.Respond.With(c, http.StatusInternalServerError, []byte("Failed to convert 'code' into a real status code number."))
        }
 
        // respond with the status
        return goweb.Respond.WithStatusText(c, statusCodeInt)
    })
 
    // /errortest should throw a system error and be handled by the
    // DefaultHttpHandler().ErrorHandler() Handler.

    goweb.Map("/errortest", func(c context.Context) error {
        return errors.New("This is a test error!")
    })
 
    /*
        Map a RESTful controller
        (see the BooksController for all the methods that will get
         mapped)
    */
    BooksController := new(BooksController)
    goweb.MapController(BooksController)
 
    goweb.Map(func(c context.Context) error {
        return goweb.API.RespondWithData(c, "Just a number!")
    }, goweb.RegexPath(`^[0-9]+$`))
 
    /*
        Catch-all handler for everything that we don't understand
    */
    goweb.Map(func(c context.Context) error {
 
        // just return a 404 message
        return goweb.API.Respond(c, 404, nil, []string{"File not found"})
 
    })
 
}
 
func main() {
 
    // map the routes
    mapRoutes()
 
    /*
 
       START OF WEB SERVER CODE
       This code is taken from Goweb server by Mat Ryer and Tyler Bunnell

    */
 
    log.Print("Goweb 2")
    log.Print("by Mat Ryer and Tyler Bunnell")
    log.Print(" ")
    log.Print("Starting Goweb powered server...")
 
    // make a http server using the goweb.DefaultHttpHandler()
    s := &http.Server{
        Addr:           Address,
        Handler:        goweb.DefaultHttpHandler(),
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
 
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    listener, listenErr := net.Listen("tcp", Address)
 
    log.Printf("  visit: %s", Address)
 
    if listenErr != nil {
        log.Fatalf("Could not listen: %s", listenErr)
    }
 
    log.Println("")
    log.Print("Point your browser to:")
    log.Printf("\t  http://localhost%s", Address)

 
    go func() {
        for _ = range c {
 
            // sig is a ^C, handle it
 
            // stop the HTTP server
            log.Print("Stopping the server...")
            listener.Close()
 
            log.Print("Tearing down...")
            log.Fatal("Finished - bye bye.  ;-)")
 
        }
    }()
 
    // begin the server
    log.Fatalf("Error in Serve: %s", s.Serve(listener))
 
    /*
 
       END OF WEB SERVER CODE
 
    */ 
}
 

// This is our Book Object
type Book struct {
    Id   string
    Title string
    Author string
    Price string
}
 
// BooksController is the RESTful MVC controller for Books.
type BooksController struct {
    // this is our fake datastore
    Books []*Book
}
 
// Before gets called before any other method.
func (r *BooksController) Before(ctx context.Context) error {
 
    // set a Books specific header
    ctx.HttpResponseWriter().Header().Set("X-Books-Controller", "true") 
    return nil
 
}
 
func (r *BooksController) Create(ctx context.Context) error {
 
    data, dataErr := ctx.RequestData()
 
    if dataErr != nil {
        return goweb.API.RespondWithError(ctx, http.StatusInternalServerError, dataErr.Error())
    }
 
    dataMap := data.(map[string]interface{})
    
    // map each value to Book Object
    Book := new(Book)
    Book.Id = dataMap["Id"].(string)
    Book.Title = dataMap["Title"].(string)
    Book.Author = dataMap["Author"].(string)
    Book.Price = dataMap["Price"].(string)    
 
    r.Books = append(r.Books, Book)
 
    return goweb.Respond.WithStatus(ctx, http.StatusCreated)
 
}
 
func (r *BooksController) ReadMany(ctx context.Context) error {
 
    if r.Books == nil {
        r.Books = make([]*Book, 0)
    }
 
    return goweb.API.RespondWithData(ctx, r.Books)
}
 
func (r *BooksController) Read(id string, ctx context.Context) error {
 
    for _, Book := range r.Books {
        // if we have a match
        if Book.Id == id {
            return goweb.API.RespondWithData(ctx, Book)
        }
    }
 
    return goweb.Respond.WithStatus(ctx, http.StatusNotFound) 
}
 
func (r *BooksController) DeleteMany(ctx context.Context) error {
 
    r.Books = make([]*Book, 0)
    return goweb.Respond.WithOK(ctx) 
}
 
func (r *BooksController) Delete(id string, ctx context.Context) error {
 
    newBooks := make([]*Book, 0)
 
    for _, Book := range r.Books {
        if Book.Id != id {
            newBooks = append(newBooks, Book)
        }
    }
 
    r.Books = newBooks
 
    return goweb.Respond.WithOK(ctx) 
}