Process Scheduling Algorithms 
This Go program implements various process scheduling algorithms including:

- First-Come, First-Served (FCFS)
- Shortest Job First (SJF)
- Shortest Job First with Priority (SJF Priority)
- Round-Robin

 ## Usage

1. **Clone the repository:**

    ```bash
    git clone <https://github.com/FQ111999/project1>
    ```

2. **Navigate to the project directory:**

    ```bash
    cd process-scheduling-algorithms
    ```

3. **Run the program:**

    ```bash
    go run main.go scheduler.go
    ```

4. **Output:**

    The program will output the schedule of processes in a Gantt chart and a table of timing for each scheduling algorithm, including average turnaround time, average waiting time, and average throughput.

## Input Format

The processes for scheduling algorithms are read from a file as the first argument to the program. Each line in the file includes a record with comma-separated fields in the following format:

