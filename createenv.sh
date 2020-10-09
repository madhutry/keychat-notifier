synapseip=`docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' synapse`
dbip=`docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' prod-pg`
echo DB_URL='postgres://keychatusr1:password@'$dbip':5432/keychatdb?sslmode=disable' > envlist.txt
echo MATRIX_URL=$synapseip:8008 >> envlist.txt
echo FRIEZE_NOTIFIER_LOG_FILE=t.txt >> envlist.txt


