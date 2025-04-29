#!/bin/bash


run_auth_command() {
  local cmd=$1
  echo "-------------------------------------------"
  echo "Running '$cmd' for auth_service..."
  echo "-------------------------------------------"
  docker-compose run --rm auth_service ./auth-service $cmd
  local exit_code=$?
  if [ $exit_code -ne 0 ]; then
    echo "-------------------------------------------"
    echo "Error: Auth service '$cmd' command failed with exit code $exit_code."
    echo "-------------------------------------------"
  else
    echo "-------------------------------------------"
    echo "Auth service '$cmd' command finished successfully."
    echo "-------------------------------------------"
  fi
  return $exit_code
}

run_user_command() {
  local cmd=$1
  echo "-------------------------------------------"
  echo "Running '$cmd' for user_service..."
  echo "-------------------------------------------"
  if [ "$cmd" == "seed" ]; then
    docker-compose run --rm user_service seed
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
      echo "-------------------------------------------"
      echo "Error: User service '$cmd' command failed with exit code $exit_code."
      echo "-------------------------------------------"
    else
      echo "-------------------------------------------"
      echo "User service '$cmd' command finished successfully."
      echo "-------------------------------------------"
    fi
    return $exit_code
  elif [ "$cmd" == "migrate" ]; then
    echo "Info: User service uses AutoMigrate on startup."
    echo "Run 'docker-compose up --build user_service' to apply migrations."
    echo "-------------------------------------------"
    return 0 # Indicate success as it's informational
  elif [ "$cmd" == "status" ]; then
     echo "Info: User service does not currently implement a 'status' command."
     echo "Migrations are checked/applied on startup."
     echo "-------------------------------------------"
     return 0
  else
    echo "Error: Command '$cmd' not explicitly supported for user_service."
    echo "-------------------------------------------"
    return 1
  fi
}

check_auth_db() {
  local db_container="auth_db"
  local db_user="kolin" # Replace if different in your docker-compose.yml
  local db_name="auth_db" # Replace if different in your docker-compose.yml
  local table_to_check="users" # Primary table for auth service

  echo "-------------------------------------------"
  echo "Checking Auth Database ($db_container)..."
  echo "-------------------------------------------"

  echo "Listing tables:"
  docker-compose exec -T $db_container psql -U $db_user -d $db_name -c "\dt"
  local exit_code_tables=$?

  if [ $exit_code_tables -ne 0 ]; then
      echo "Error checking tables for $db_container."
  else
      echo "Checking row count for '$table_to_check' table:"
      docker-compose exec -T $db_container psql -U $db_user -d $db_name -c "SELECT COUNT(*) FROM $table_to_check;"
      local exit_code_count=$?
      if [ $exit_code_count -ne 0 ]; then
          echo "Error checking row count for '$table_to_check' in $db_container."
      fi
  fi
  echo "-------------------------------------------"
  echo "Auth Database check finished."
  echo "-------------------------------------------"
  return $((exit_code_tables + exit_code_count))
}

check_user_db() {
  local db_container="user_db"
  local db_user="kolin" # From pkg/db/db.go default
  local db_name="user_db" # From pkg/db/db.go default
  local table_to_check="users" # Primary table for user service

  echo "-------------------------------------------"
  echo "Checking User Database ($db_container)..."
  echo "-------------------------------------------"

  echo "Listing tables:"
  docker-compose exec -T $db_container psql -U $db_user -d $db_name -c "\dt"
  local exit_code_tables=$?

  if [ $exit_code_tables -ne 0 ]; then
      echo "Error checking tables for $db_container."
  else
      echo "Checking row count for '$table_to_check' table:"
      docker-compose exec -T $db_container psql -U $db_user -d $db_name -c "SELECT COUNT(*) FROM $table_to_check;"
      local exit_code_count=$?
      if [ $exit_code_count -ne 0 ]; then
          echo "Error checking row count for '$table_to_check' in $db_container."
      fi
  fi
  echo "-------------------------------------------"
  echo "User Database check finished."
  echo "-------------------------------------------"
  return $((exit_code_tables + exit_code_count))
}

run_single_service_check() {
    local service_display_name=$1
    local db_container=$2
    local db_user=$3
    local db_name=$4

    echo "--- Comprehensive Check: $service_display_name ---"

    local container_id=$(docker-compose ps -q $db_container)
    if [ -z "$container_id" ]; then
        echo " [FAIL] Database container '$db_container' is not running."
        echo "-------------------------------------------"
        return 1 # Stop checks for this service if DB isn't running
    else
        echo " [OK]   Database container '$db_container' is running."
    fi

    local tables=$(docker-compose exec -T $db_container psql -U $db_user -d $db_name -tAc "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';")
    if [ $? -ne 0 ]; then
        echo " [FAIL] Migration check failed (error querying tables)."
        echo "-------------------------------------------"
        return 1
    elif [ -z "$tables" ]; then
        echo " [FAIL] Migration check failed (no user tables found in database '$db_name')."
        echo "-------------------------------------------"
        return 1
    else
        echo " [OK]   Migration check passed (found user tables in '$db_name')."
    fi

    echo " [INFO] Checking row counts and sample data for all user tables:"
    local all_tables_empty=true
    local query_errors=false
    for table in $tables; do
        local row_count=$(docker-compose exec -T $db_container psql -U $db_user -d $db_name -tAc "SELECT COUNT(*) FROM $table;")
        if [ $? -eq 0 ]; then
            echo "        - Table '$table': $row_count rows"
            if [ "$row_count" -gt 0 ]; then
                all_tables_empty=false
                echo "          Sample Data (first 3 rows):"
                docker-compose exec -T $db_container psql -U $db_user -d $db_name -c "SELECT * FROM \"$table\" LIMIT 3;"
                local sample_exit_code=$?
                if [ $sample_exit_code -ne 0 ]; then
                    echo "          Error fetching sample data for table '$table'."
                fi
            fi
        else
            echo "        - Table '$table': Error querying row count."
            query_errors=true
        fi
    done

    if [ "$query_errors" = true ]; then
         echo " [WARN] Seeding check encountered errors querying row counts for one or more tables."
    elif [ "$all_tables_empty" = true ]; then
        echo " [WARN] Seeding check: All user tables exist but are empty."
    else
        echo " [OK]   Seeding check passed (at least one table contains data)."
    fi


    echo "-------------------------------------------"
    return 0 # Return success overall for this service check completion
}

run_comprehensive_check() {
    echo "==========================================="
    echo " Starting Comprehensive Checks..."
    echo "==========================================="

    run_single_service_check "Auth Service" "auth_db" "kolin" "auth_db" # Removed "users"
    run_single_service_check "User Service" "user_db" "kolin" "user_db" # Removed "users"

    echo "==========================================="
    echo " Comprehensive Checks Finished."
    echo "==========================================="
}

show_command_menu() {
  local service_name=$1
  local service_display_name=$2
  while true; do
    clear
    echo "==========================================="
    echo " Manage Service: $service_display_name"
    echo "==========================================="
    echo " 1) Migrate Database"
    echo " 2) Seed Database"
    echo " 3) Check Migration Status"
    echo " 4) Check Database (Tables & Users Count)"
    echo " -----------------------------------------"
    echo " b) Back to Main Menu"
    echo "==========================================="
    read -p "Enter command choice: " cmd_choice

    case $cmd_choice in
      1)
        if [ "$service_name" == "auth" ]; then
          run_auth_command "migrate"
        elif [ "$service_name" == "user" ]; then
          run_user_command "migrate"
        fi
        read -p "Press Enter to continue..."
        ;;
      2)
        if [ "$service_name" == "auth" ]; then
          run_auth_command "seed"
        elif [ "$service_name" == "user" ]; then
          run_user_command "seed"
        fi
        read -p "Press Enter to continue..."
        ;;
      3)
        if [ "$service_name" == "auth" ]; then
          run_auth_command "status"
        elif [ "$service_name" == "user" ]; then
          run_user_command "status"
        fi
         read -p "Press Enter to continue..."
         ;;
      4)
        if [ "$service_name" == "auth" ]; then
          check_auth_db
        elif [ "$service_name" == "user" ]; then
          check_user_db
        fi
         read -p "Press Enter to continue..."
         ;;
      b|B)
        break # Exit command menu loop
        ;;
      *)
        echo "Invalid command choice. Please try again."
        read -p "Press Enter to continue..."
        ;;
    esac
  done
}


show_all_services_menu() {
    while true; do
        clear
        echo "==========================================="
        echo " Manage All Services"
        echo "==========================================="
        echo " 1) Migrate All"
        echo " 2) Seed All"
        echo " 3) Check All Databases"
        echo " 4) Run Comprehensive Check"
        echo " -----------------------------------------"
        echo " b) Back to Main Menu"
        echo "==========================================="
        read -p "Enter command choice: " cmd_choice

        case $cmd_choice in
            1)
                run_auth_command "migrate"
                run_user_command "migrate"
                read -p "Press Enter to continue..."
                ;;
            2)
                run_auth_command "seed"
                run_user_command "seed"
                read -p "Press Enter to continue..."
                ;;
            3)
                check_auth_db
                check_user_db
                read -p "Press Enter to continue..."
                ;;
            4)
                run_comprehensive_check
                read -p "Press Enter to continue..."
                ;;
            b|B)
                break # Exit command menu loop
                ;;
            *)
                echo "Invalid command choice. Please try again."
                read -p "Press Enter to continue..."
                ;;
        esac
    done
}


while true; do
  clear # Clear the screen for a clean menu
  echo "==========================================="
  echo " Backend Service Management Menu"
  echo "==========================================="
  echo " 1) Auth Service"
  echo " 2) User Service"
  echo " 3) All Services (Migrate/Seed/Check DB/Comprehensive Check)"
  echo " -----------------------------------------"
  echo " q) Quit"
  echo "==========================================="
  read -p "Choose an option: " main_choice

  case $main_choice in
    1)
      show_command_menu "auth" "Auth Service"
      ;;
    2)
      show_command_menu "user" "User Service"
      ;;
    3)
      show_all_services_menu
      ;;
    q|Q)
      echo "Exiting script."
      break # Exit the main loop
      ;;
    *)
      echo "Invalid option. Please try again."
      read -p "Press Enter to continue..."
      ;;
  esac
done

exit 0
