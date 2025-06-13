import os
import numpy as np


def calculate_average_from_files(directory_path):
    """
    Calculates the average of integers from each file in a directory.

    Args:
        directory_path: The path to the directory containing the files.
    """
    if not os.path.isdir(directory_path):
        print(f"Error: Directory not found at '{directory_path}'")
        return

    for filename in os.listdir(directory_path):
        filepath = os.path.join(directory_path, filename)

        if os.path.isfile(filepath):
            try:
                with open(filepath, 'r') as f:
                    lines = f.readlines()
                    numbers = [int(line.strip()) for line in lines]

                    warmup_requests = int(len(numbers) * 0.15)
                    numbers = numbers[warmup_requests:]

                    if numbers:
                        average = sum(numbers) / len(numbers)
                        p99_value = np.percentile(numbers, 99)
                        print(f"{filename}: {average}")
                        # print(
                        #     f"{filename}: average={average},\tp99={p99_value}")

                    else:
                        print(f"{filename}: The file is empty.")

            except FileNotFoundError:
                print(f"Error: File not found at '{filepath}'")
            except ValueError:
                print(
                    f"Error: Could not convert data to an integer in {filename}."
                )
            except Exception as e:
                print(f"An unexpected error occurred with {filename}: {e}")


if __name__ == '__main__':
    target_directory = 'message_sizes'
    calculate_average_from_files(target_directory)
