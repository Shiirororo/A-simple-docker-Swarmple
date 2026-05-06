import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
  vus: 600,         
  duration: '300s',  
  noVUConnectionReuse: true,
};

export default function () {
  http.get('http://localhost:80/count');
  sleep(1);
}