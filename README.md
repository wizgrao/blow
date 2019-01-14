# Blow
MapReduce inspired distributeed computing framework written in Go. Just define a data source, mapping functions, and a connection between a host and worker computers.

## Example
We tackle a popular problem, fizzbuzz! Fizzbuzz, if you're not acquainted with the problem, asks you to count up to a certain number and print "fizz" if the number is divisible by 3, "buzz" if its divisible by 5, "fizzbuzz" if divisible by 15, or just the number otherwise.

This is a good sample problem to tackle since essentially we are mapping numbers to their phrase independantly, so it makes sense to distribute this problem to worker computers. All of the code can be found in [The fizzbuzz folder](cmd/fizzbuzz)

### Data Types
Our data types have to implement the Keyed interface, meaning that they must have a `Key()` function that returns an int which should be some sort of hash. Our map would go from integers to phrases, but the output struct should also contain the number for which the phrase corresponds to. So, we have a FizzyInput struct which just contains an Int val, and a FizzBuzz struct that contains the int val and the phrase as well. These are defined in the [protobuf file](cmd/fizzbuzz/fizz.proto) and implemented in [the fizzbuzz.go file](cmd/fizzbuzz/fizzbuzz.go).
```golang
func (f *FizzyInput) Key() int {
	return int(f.Val)
}

func (f *FizzBuzz) Key() int {
	return int(f.Number)
}
```
In order to send these over a channel, we want to define how to Marshal and UnMarshal these values. We do this by implementing the `maps.Encoder` interface
```golang
type FizzyInputMarshaller struct {
}

func (*FizzyInputMarshaller) Marshal(k maps.Keyed) ([]byte, error) {
	return proto.Marshal(k.(*FizzyInput))
}
func (*FizzyInputMarshaller) UnMarshal(b []byte) (maps.Keyed, error) {
	o := &FizzyInput{}
	err := proto.Unmarshal(b, o)
	return o, err
}

type FizzBuzzMarshaller struct {
}

func (*FizzBuzzMarshaller) Marshal(k maps.Keyed) ([]byte, error) {
	return proto.Marshal(k.(*FizzBuzz))
}
func (*FizzBuzzMarshaller) UnMarshal(b []byte) (maps.Keyed, error) {
	o := &FizzBuzz{}
	err := proto.Unmarshal(b, o)
	return o, err
}
```
We use protocol buffers here since they are very fast, but we could use any encoding scheme we want here, such as `json`.

### Data Source
We want to implement the `maps.Source` interface. For this, we just output the numbers we want to find the associated phrase to. 
```golang
type FizzGenerator int

func (FizzGenerator) Do(c chan <- maps.Keyed) {
	for i:=0; i< 10000; i++ {
		c <- &FizzyInput{Val: int32(i)}
	}
}
```
### Map
Now we implement the map itself, which needs the `Do` method, an `ID` for the map itself, and functions that return an encoder for its in values, and an encoder for its out values.
```golang
type FizzMapper int

func (FizzMapper) Do(k maps.Keyed, c chan <- maps.Keyed) {
	n := k.(*FizzyInput).Val
	time.Sleep(time.Second/4)
	if n % 15 == 0 {
		c <- &FizzBuzz{
			Number: n,
			Word:   "fizzbuzz",
		}

	} else if n % 3 == 0 {
		c <- &FizzBuzz{
			Number: n,
			Word:   "fizz",
		}
	} else if n % 5 == 0 {
		c <- &FizzBuzz{
			Number: n,
			Word:   "buzz",
		}
	} else {
		c <- &FizzBuzz{
			Number: n,
			Word:   strconv.FormatInt(int64(n), 10),
		}
	}
}

func (FizzMapper) ID() string{
	return "fizzymapper"
}

func (f FizzMapper) InEncoder() maps.Encoder {
	return &FizzyInputMarshaller{}
}
func (f FizzMapper) OutEncoder() maps.Encoder {
	return &FizzBuzzMarshaller{}
}
```
Here we also add a sleep to simulate a harder task being accomplished to warrent this parallelization


### The server
Now we need a way to connect to the workers. Here, our workers are going to be connected browsers that run code through the webassembly target. We can use websockets, and the existing Gorilla connection and Browser Websocket connection wrappers that are included in this framework. We define a web handler that turns a gorilla websocket into a connection, and adds it to the worker pool. 
```golang
var pool *maps.WorkerPool
type newConnectionHandler int

func (newConnectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err !=nil {
		fmt.Println("Error upgrading http", err)
		return
	}
	connection := &gorillaconnection.Connection{c}
	fmt.Println("New Connection")
	pool.AddWorker(connection)
	select {}
}
```
### The client
This is the code that would run on the browser. We have to create an instance of the mappers we want the client to be able to handle, register it with the host, and start waiting for requests.
```golang
func main() {
	sock := wasmsocket.GetSocket("socket")
	var mapper fizzbuzz.FizzMapper
	h := maps.NewHost(sock)
	h.Register(mapper)
	h.Start()
	select {}
}
```
### The pipeline
Now, on the server, we have to actually define the mapping pipeline we want to use. We source using the previously defined source from above, we then map it to the network with the mapper, we the use the built in print mapper to print out the results, and then discard the data.
Then, we register the mapper to the worker pool and serve the client and the compiled client files. 
```golang
func main() {
	var nch newConnectionHandler
	pool = maps.NewWorkerPool()
	pool.Register(fizzmapper)
	go func() {
		maps.GeneratorSource(generator, pool).MapDispatch(fizzmapper).MapLocalParallel(&maps.PrintMapper{}, 10).Sink()
	}()
	http.Handle("/main.wasm", wasm)
	http.Handle("/sock", nch)
	http.Handle("/", http.FileServer(http.Dir("slave/")))
	fmt.Print(http.ListenAndServe(":8090", nil))
}
```
Run the server, open a browser to `localhost:8090` and you'll have your own distributed fizzbuzz app!


