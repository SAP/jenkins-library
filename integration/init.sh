while read p; do
  docker pull "$p"
done <images.txt
