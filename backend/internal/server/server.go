package server

func main() {

}

nc serve() {
	var wait time.duration
	flag.durationvar(&wait, "graceful-timeout", time.second*15, "duration to wait for existing connections to close")
	flag.parse()

	r := mux.newrouter()
	r.handlefunc("/signin", signin).methods(http.methodpost, http.methodoptions)
	r.handlefunc("/api/linktoken", createlinktoken).methods(http.methodpost, http.methodoptions)
	r.handlefunc("/api/publictoken", exchangepublictoken).methods(http.methodpost, http.methodoptions)
	r.handlefunc("/api/transactions", gettransactions).methods(http.methodget, http.methodoptions)
	r.handlefunc("/api/accounts", getallaccounts).methods(http.methodget, http.methodoptions)

	r.use(mux.corsmethodmiddleware(r))
	r.use(corsmiddleware)

	srv := &http.server{
		addr:         "0.0.0.0:8080",
		writetimeout: time.second * 15,
		readtimeout:  time.second * 15,
		idletimeout:  time.second * 60,
		handler:      r,
	}
	go func() {
		if err := srv.listenandserve(); err != nil {
			log.println(err)
		}
	}()

	c := make(chan os.signal, 1)

	signal.notify(c, os.interrupt)

	<-c

	ctx, cancel := context.withtimeout(context.background(), wait)
	defer cancel()

	srv.shutdown(ctx)

	log.println("shutting down")
	os.exit(0)
}

func init()
