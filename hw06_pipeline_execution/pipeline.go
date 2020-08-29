package hw06_pipeline_execution //nolint:golint,stylecheck

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	// Place your code here
	doneChanel := make(Bi)
	go func() {
		<-done
		close(doneChanel)
	}()
	out := in
	for _, stage := range stages {
		newIn := make(Bi)
		go func(_in Bi, _out Out) {
			defer close(_in)
			for {
				select {
				case <-doneChanel:
					return
				case v, ok := <-_out:
					if !ok {
						return
					}
					_in <- v
				}
			}
		}(newIn, out)
		out = stage(newIn)
	}

	return out
}
