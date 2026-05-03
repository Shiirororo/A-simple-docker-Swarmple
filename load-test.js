import http from 'k6/http';
import { sleep } from 'k6';

export let options = {
  vus: 50,         
  duration: '100s',  
};

export default function () {
  http.get('http://localhost:80/count');
  sleep(1);
}