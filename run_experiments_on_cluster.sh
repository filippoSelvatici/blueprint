#!/bin/bash

remote="USERNAME@HOSTNAME"
remote_blueprint_path="$remote:/home/USERNAME/blueprint"
workload_path="examples/sockshop/cmplx_workload/workloadgen/workload.go"

replace_method_proportion() {
  local method_name="$1"
  local new_value="$2"

  sed -i "s/\(w\.Run$method_name, \)[0-9]\+/\1$new_value/" "$workload_path"
}

replace_throughput() {
  local new_value="$1"

  sed -i "s/\(var tput = flag.Int64(\"tput\", \)[0-9]\+/\1$new_value/" "$workload_path"
}

run_experiment() {
  local experiment_name="$1"
  local throughput="$2"
  local wiring="$3"
  local sockshop_path="examples/sockshop"
  local path_of_stats="$sockshop_path/build/stats.csv"
  local experiment_folder="experiments/$experiment_name/$throughput"

  mkdir -p "$experiment_folder"

  scp $workload_path $remote_blueprint_path/$workload_path > /dev/null
  ssh $remote 'cd blueprint && bash run_experiments.sh '$wiring'' > "$experiment_folder/$wiring-ssh-logs.txt" 2>&1
  scp $remote_blueprint_path/$path_of_stats $experiment_folder/$wiring.csv > /dev/null
  scp $remote_blueprint_path/$sockshop_path/wllogs.txt $experiment_folder/$wiring-wllogs.txt > /dev/null
  scp $remote_blueprint_path/$sockshop_path/logs.txt $experiment_folder/$wiring-logs.txt > /dev/null
}

# You should provide proportions for all the methods, otherwise you are relying on the values that were there before.
# The format is:
# '(["GetCart"]=N ["AddItem"]=N ["ListItems"]=N ["ListTags"]=N ["Register"]=N ["PostCard"]=N ["PostAddress"]=N ["AddItemToSpecificUser"]=N ["NewOrder"]=N ["AddItemAndNewOrder"]=N)'
declare -A experiments=(
  ["only_list_items"]='(["GetCart"]=0 ["AddItem"]=0 ["ListItems"]=100 ["ListTags"]=0 ["Register"]=0 ["PostCard"]=0 ["PostAddress"]=0 ["AddItemToSpecificUser"]=0 ["NewOrder"]=0 ["AddItemAndNewOrder"]=0)'
  ["only_post_methods"]='(["GetCart"]=0 ["AddItem"]=0 ["ListItems"]=0 ["ListTags"]=0 ["Register"]=0 ["PostCard"]=50 ["PostAddress"]=50 ["AddItemToSpecificUser"]=0 ["NewOrder"]=0 ["AddItemAndNewOrder"]=0)'
  ["new_orders"]='(["GetCart"]=0 ["AddItem"]=0 ["ListItems"]=0 ["ListTags"]=0 ["Register"]=0 ["PostCard"]=0 ["PostAddress"]=0 ["AddItemToSpecificUser"]=70 ["NewOrder"]=30 ["AddItemAndNewOrder"]=0)'
  ["all_types"]='(["GetCart"]=15 ["AddItem"]=10 ["ListItems"]=15 ["ListTags"]=15 ["Register"]=4 ["PostCard"]=2 ["PostAddress"]=2 ["AddItemToSpecificUser"]=27 ["NewOrder"]=10 ["AddItemAndNewOrder"]=0)'
)

for experiment_name in "${!experiments[@]}"; do
  experiment="${experiments[$experiment_name]}"
  printf "######################\nEXPERIMENT: $experiment_name\n######################\n"
  declare -A method_proportions="${experiment}"
  for method in "${!method_proportions[@]}"; do
    replace_method_proportion "$method" "${method_proportions[$method]}"
  done

  # TODO update these with the desired RPS values. Note that the application can handle different RPS values
  # depending on the experiment 
  for throughput in 1; do 
    printf "> Running experiment with throughput: $throughput\n"
    replace_throughput $throughput

    printf ">> Running namedpipes experiment\n"
    run_experiment $experiment_name $throughput "namedpipes"

    printf ">> Running docker (gRPC) experiment\n"
    run_experiment $experiment_name $throughput "docker"
    
    printf "\n\n"
  done
done

printf "All experiments completed.\n"
