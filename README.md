# gomultitask

Simple wrapper for go multi task services.
U can use it for run some root goroutines from main file.
For example: http server, async worker system or additional info server.
Simple implement `task.Interface` for your routine.
U can see example in tests.

System support exit by err from one of tasks or signals.
If you set custom  `FallNumber` in task config, for example 3, 
your task will be restart after 3 falls,
but next time err from task will stop application.

If system catch panic, application will be stopped immediately 
with graceful shutdown another tasks.

If You find any errors in code or want improvement,
please write issue with tag `bug` or `feature`.  

## Version
0.0.1
