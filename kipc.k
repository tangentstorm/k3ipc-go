/ k-ipc server for kipc_ws.py

/ schedule: do task as soon as possible after given timestamp
/ .tasks: ,{`0:,"timer is running."}
/ .times: ,_t

/ set up a timer to allow deferred responses
/ \t 1
/ .m.ts: {
/   due: .times < _t
/   {x()}' .tasks @ & due
/   .times@: & ~due
/   .tasks@: & ~due }

/ listen for synchronous requests, respond with result of eval
.m.g:{
  `0:,".m.g[",(5:x),"]"             // print msg we got to stdout
  5: . x}

/ listen for asynchronous requests, respond with result of eval
.m.s:{
  `0:,".m.s[",(5:x),"]"             // print msg we got to stdout
  tx: x 0                           // transaction id
  k3: x 1                           // code to evaluate
  / r: @[.:;k3;:];                    // evaluate it with error trap
  r: . k3
  `0: , 5: r
  / TODO: get the callback working for kona
  / h: 3: (`"0.0.0.0";1024)
  / h 3: tx
  /.times,: _t + * 1 _draw 5         // send response in a few seconds
  /.tasks,: {[h;r;tx;x] `0:,5:tx,r; h 3: tx,r }[h;r;tx]
}
