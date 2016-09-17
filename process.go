package main

import (
	"log"
	"time"
)

func HandleRunning(g *Global, status LaunchStatus) {
	log.Println("Process", status.Name, "running")

	var proc *Process = g.Procs[status.Name]
	proc.Running = true
}

/* Called whenever a child exits. Take appropriate action, such as restarting.
 */
func HandleDone(g *Global, status LaunchStatus) {
	var proc *Process = g.Procs[status.Name]

	/* If there was an error and we should try to start it again. */
	if status.Err != nil && proc.Config.IgnoreFailure == false {
		log.Println("Process", proc.Config.Name, "failed after",
			status.Duration.String())
		/* Give up if it failed too quickly. */
		if proc.Config.MinRuntime != 0 &&
			time.Duration(proc.Config.MinRuntime)*time.Millisecond >
				status.Duration {
			log.Println("Process", proc.Config.Name,
				"failed too quickly. Giving up.")
			/* If it didn't fail too quickly, continue with restart. */
		} else {
			/* Wait the required time before restarting. */
			time.Sleep(time.Duration(proc.Config.RestartDelay) * time.Millisecond)

			/* Actually restart it. */
			log.Println("Process", proc.Config.Name, "launching")
			go LaunchProcess(proc.Config, g)
			g.RunningProcesses++
		}

		/* If the process completed successfully or we don't care. */
	} else {
		log.Println("Process", proc.Config.Name, "finished after",
			status.Duration.String())
	}
}
