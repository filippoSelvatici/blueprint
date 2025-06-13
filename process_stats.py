# read the content of stats.csv and compute the 50th, 75th, 90th, 95th, 99th, 99.9th percentiles
import os
import pandas as pd
import pprint
import matplotlib.pyplot as plt
import numpy as np

#################### Percentiles ####################


def read_stats(file_path):
    """Read the stats.csv file and return a list of DataFrames, one for each method."""
    dfs = []
    with open(file_path, 'r') as file:
        lines = file.readlines()

    current_method = None
    data = []
    for line in lines:
        line = line.strip()
        if line.startswith("METHOD"):
            if current_method and data:
                # Create a DataFrame for the previous method
                df = pd.DataFrame(data,
                                  columns=["Start", "Duration", "IsError"])
                dfs.append({"method": current_method, "df": df})
                data = []
            current_method = line.split(" ", 1)[1]  # Extract method name
        elif line and not line.startswith("Start"):
            # Parse the data rows
            parts = line.split(",")
            if len(parts) == 3:
                data.append(
                    [float(parts[0]),
                     float(parts[1]), parts[2] == "true"])

    # Add the last method's data
    if current_method and data:
        df = pd.DataFrame(data, columns=["Start", "Duration", "IsError"])
        dfs.append({"method": current_method, "df": df})

    return dfs


def drop_warmup_iterations(dfs, percentage=15):
    """Drop the first percentage% of rows from each DataFrame to remove warmup iterations."""
    for dfAndMethodName in dfs:
        df = dfAndMethodName["df"]
        n = len(df)
        if n > percentage:  # Ensure there are enough rows to drop
            df.drop(df.index[:int(n * percentage / 100)], inplace=True)
    return dfs


def get_total_df(dfs):
    """Combine all DataFrames into a single DataFrame."""
    total_df = pd.DataFrame()
    for dfAndMethodName in dfs:
        method = dfAndMethodName["method"]
        df = dfAndMethodName["df"]
        df["Method"] = method  # Add a column for the method name
        total_df = pd.concat([total_df, df], ignore_index=True)
    return total_df


def compute_percentiles(df):
    percentiles = [50, 75, 90, 95, 99, 99.9]
    return {f"{p}th": df["Duration"].quantile(p / 100) for p in percentiles}


def compute_average(df):
    return df["Duration"].mean()


def compute_percentiles_for_all_methods(dfs):
    return {
        dfAndMethodName["method"]: compute_percentiles(dfAndMethodName["df"])
        for dfAndMethodName in dfs
    }


def get_method_proportions(df):
    proportions = {}
    for _, row in df.iterrows():
        if row["Method"] not in proportions:
            proportions[row["Method"]] = 0
        proportions[row["Method"]] += 1
    for (key, val) in proportions.items():
        proportions[key] = float(int(val * 10000 / len(df)) / 100.0)
    return proportions


################### Plot Latencies over Time ###################


def plot_request_durations(df,
                           execution_id_col='ExecutionID',
                           start_col='Start',
                           duration_col='Duration'):
    """
    Plots request durations over time for different executions from a pandas DataFrame.

    Args:
        df (pd.DataFrame): DataFrame containing the data.
                           Must include columns for start time, duration, and execution ID.
        execution_id_col (str): Name of the column identifying different executions.
        start_col (str): Name of the column for start times (in nanoseconds).
        duration_col (str): Name of the column for durations (in nanoseconds).
    """
    if df.empty:
        print("DataFrame is empty. Nothing to plot.")
        return

    if not all(col in df.columns
               for col in [start_col, duration_col, execution_id_col]):
        print(
            f"Error: DataFrame must contain columns: '{start_col}', '{duration_col}', and '{execution_id_col}'."
        )
        print(f"Available columns: {df.columns.tolist()}")
        return

    # Make a copy to avoid modifying the original DataFrame
    plot_df = df.copy()

    # Convert start and duration times
    # Find the overall minimum start time to use as t=0
    min_overall_start_ns = plot_df[start_col].min()

    # Calculate time since the overall start in seconds
    plot_df['TimeSinceStart_s'] = (plot_df[start_col] -
                                   min_overall_start_ns) / 1e9

    # Convert duration to milliseconds
    plot_df['Duration_ms'] = plot_df[duration_col] / 1e6

    # Determine the x-axis limit: up to the start of the last request
    if not plot_df.empty:
        max_start_time_s = (plot_df[start_col].max() -
                            min_overall_start_ns) / 1e9
    else:
        max_start_time_s = 0

    plt.figure(figsize=(12, 7))

    # Plot a line for each execution
    for execution_id in sorted(plot_df[execution_id_col].unique()):
        execution_data = plot_df[plot_df[execution_id_col] ==
                                 execution_id].sort_values(
                                     by='TimeSinceStart_s')
        if not execution_data.empty:
            plt.plot(
                execution_data['TimeSinceStart_s'],
                execution_data['Duration_ms'],
                marker='o',  # Add markers to see individual points
                linestyle='-',
                label=f'Execution {execution_id}')

    plt.xlabel("Time Since Start (seconds)")
    plt.ylabel("Duration (ms)")
    plt.yscale('log')  # Set y-axis to exponential/logarithmic scale
    plt.title("Request Durations Over Time by Execution")

    # Set x-axis limit
    # Add a small padding to the max_start_time_s if there are data points,
    # otherwise, matplotlib will auto-scale. If max_start_time_s is 0 (e.g. one event),
    # give it some positive range.
    if max_start_time_s > 0:
        plt.xlim(0, max_start_time_s * 1.05)  # 5% padding
    elif not plot_df.empty:  # Single event or all events start at the same time
        plt.xlim(0, 1)  # Default range if only one time point at 0
    else:  # Empty dataframe after filtering
        plt.xlim(0, 1)

    if plot_df[execution_id_col].nunique(
    ) > 1 or not plot_df.empty:  # Add legend if there are lines to label
        plt.legend()

    plt.grid(True, which="both", ls="--",
             alpha=0.7)  # Add grid for better readability
    plt.tight_layout(
    )  # Adjust plot to ensure everything fits without overlapping


folders = [
    # "experiments-final/all_types/64",
    # "experiments-final/all_types/128",
    # "experiments-final/new_orders/1024",
    # "experiments-final/new_orders/1536",
    # "experiments-final/new_orders/2048",
    # "experiments-final/new_orders/2560",
    # "experiments-final/new_orders/3072",
    # "experiments-final/only_list_items/1024",
    # "experiments-final/only_list_items/1280",
]

for folder in folders:
    print(f"########## {folder} ##########")
    print(f"[namedpipes] latency percentiles:")
    pprint.pp(
        compute_percentiles_for_all_methods(
            drop_warmup_iterations(read_stats(f"{folder}/namedpipes.csv"))))

    plot_request_durations(get_total_df(
        read_stats(f"{folder}/namedpipes.csv")),
                           execution_id_col='Method')

    print(f"[gRPC] latency percentiles:")
    pprint.pp(
        compute_percentiles_for_all_methods(
            drop_warmup_iterations(read_stats(f"{folder}/docker.csv"))))

    plot_request_durations(get_total_df(read_stats(f"{folder}/docker.csv")),
                           execution_id_col='Method')

    plt.show()
    print("\n\n")

# given a folder name, look into all the folders inside it. You can assume that the names of the folders inside are just numbers representing RPS. Inside the RPS folders, there is a docker.csv and namedpipes.csv files. Read

foldersp99 = [
    ("experiments-final/all_types", 170),
    ("experiments-final/new_orders", 2600),
    ("experiments-final/only_list_items", 1800),
    ("experiments-final/only_post_methods", 150),
]

for (folder, max_rps_to_plot) in foldersp99:
    rpss = os.listdir(folder)
    print("RPSs available:", rpss)

    namedpipes_p99 = {}
    docker_p99 = {}
    docker_single_ctr_p99 = {}
    for rps in rpss:
        # for both docker and namedpipes
        # - read csv
        # - drop warmup
        # - get_total_df
        # - compute_percentiles
        # - add p99 percentile to dictionaries
        if int(rps) > max_rps_to_plot:
            continue

        namedpipes_df = get_total_df(
            drop_warmup_iterations(
                read_stats(f"{folder}/{rps}/namedpipes.csv")))

        # namedpipes_p99[rps] = compute_average(namedpipes_df)
        namedpipes_p99[rps] = compute_percentiles(namedpipes_df)["99th"]

        ###################################################

        docker_df = get_total_df(
            drop_warmup_iterations(read_stats(f"{folder}/{rps}/docker.csv")))

        # docker_p99[rps] = compute_average(docker_df)
        docker_p99[rps] = compute_percentiles(docker_df)["99th"]

    print("namedpipes: ", namedpipes_p99)
    print("docker: ", docker_p99)
    # Plot the p99 percentiles for namedpipes and docker
    plt.figure(figsize=(10, 6))

    # Convert RPSs to integers for sorting and plotting
    rpss_int = list(map(int, namedpipes_p99.keys()))
    rpss_int.sort()

    # Plot namedpipes p99
    plt.plot(rpss_int,
             [namedpipes_p99[str(rps)] / 1000000 for rps in rpss_int],
             marker='o',
             linestyle='-',
             label='Named Pipes')

    # Plot docker p99
    plt.plot(rpss_int, [docker_p99[str(rps)] / 1000000 for rps in rpss_int],
             marker='o',
             linestyle='-',
             label='gRPC')

    plt.xlabel("Requests Per Second (RPS)")
    plt.ylabel("p99 Latency (ms)")
    plt.title(
        f"p99 Latency vs RPS for Namedpipes and gRPC - experiment {folder}")
    plt.legend()
    plt.grid(True, which="both", ls="--", alpha=0.7)
    plt.tight_layout()

plt.show()
