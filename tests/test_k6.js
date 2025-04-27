import http from 'k6/http';
import { check } from 'k6';

export let options = {
    vus: 1000,          // 1000 виртуальных пользователей
    iterations: 5000,   // всего 5000 запросов
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% запросов быстрее 500мс
    },
};

export default function () {
    let res = http.get('http://localhost:8080/');

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
    });
}
