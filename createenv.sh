synapseip=`docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' synapse`
dbip=`docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' dev-postgres`
echo DB_URL='postgres://***REMOVED***:***REMOVED***@'$dbip':5432/friezechat?sslmode=disable' > envlist.txt
echo MATRIX_URL=$synapseip:8008 >> envlist.txt
echo FRIEZE_NOTIFIER_LOG_FILE=t.txt >> envlist.txt


