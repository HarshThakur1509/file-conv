package routes

import (
	"file-conv/internal/handlers"
	"file-conv/internal/middleware"
	"fmt"
	"net/http"

	"github.com/rs/cors"
)

type ApiServer struct {
	addr string
}

func NewApiServer(addr string) *ApiServer {
	return &ApiServer{addr: addr}
}

func (s *ApiServer) Run() error {
	router := http.NewServeMux()

	router.HandleFunc("POST /convert/jpg-to-png", handlers.ConvertJPGToPNG)
	router.HandleFunc("POST /convert/png-to-jpg", handlers.ConvertPNGToJPG)
	router.HandleFunc("POST /convert/to-pdf", handlers.ConvertToPDF)
	router.HandleFunc("POST /compress", handlers.CompressImage)
	router.HandleFunc("POST /resize", handlers.ResizeImage)
	router.HandleFunc("POST /transparent", handlers.BackgroundTransparent)

	router.HandleFunc("POST /merge-pdfs", handlers.MergePDFs)
	router.HandleFunc("POST /split-pdf", handlers.SplitPDF)
	// Add code here

	stack := middleware.MiddlewareChain(middleware.Logger, middleware.RecoveryMiddleware)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // Specify your React frontend origin
		AllowCredentials: true,                              // Allow cookies and credentials
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"Accept",
			"Origin",
			"X-Requested-With"},
	}).Handler(stack(router))

	server := http.Server{
		Addr:    s.addr,
		Handler: corsHandler,
	}
	fmt.Println("Server has started", s.addr)
	return server.ListenAndServe()
}
