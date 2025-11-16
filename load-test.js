import http from 'k6/http';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
    scenarios: {
        steady_load: {
            executor: 'constant-arrival-rate',
            rate: 5,
            timeUnit: '1s',
            duration: '5m',
            preAllocatedVUs: 10,
            maxVUs: 50,
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<300'],
    },
};

export function setup() {
    return {}; // setup не нужен
}

export default function () {
    const endpoints = [
        { method: 'GET',  url: `${BASE_URL}/team/get?team_name=team-1` },
        { method: 'POST', url: `${BASE_URL}/users/setIsActive`, body: JSON.stringify({ user_id: "u1", is_active: true }) },
        { method: 'GET',  url: `${BASE_URL}/users/getReview?user_id=u1` },
    ];

    const req = endpoints[Math.floor(Math.random() * endpoints.length)];

    if (req.method === 'GET') {
        http.get(req.url);
    } else {
        http.post(req.url, req.body, { headers: { 'Content-Type': 'application/json' } });
    }
}
