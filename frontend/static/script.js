let firstExpressionSent = false;
let expressionsIntervalId = null;
let resultIntervalId = null;

// Инициализация
document.addEventListener('DOMContentLoaded', async () => {

    // Инициализация темы
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'dark') {
        document.body.classList.add('dark-theme');
    }

    // Инициализация элементов
    document.getElementById('themeToggle').addEventListener('click', toggleTheme);
    document.getElementById('calculateBtn').addEventListener('click', sendExpression);
    document.getElementById('expression').addEventListener('keydown', handleEnterKey);

    // Загрузка истории
    await fetchExpressions();
});

// Обертка для запросов с авторизацией
async function fetchWithAuth(url, options = {}) {
    try {
        const response = await fetch(url, {
            ...options,
            credentials: 'include'
        });

        if (response.status === 401) {
            const refreshResponse = await fetch('http://localhost:8080/api/v1/refresh-token', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include'
            });

            if (!refreshResponse.ok) {
                window.location.href = 'auth/login.html';
                throw new Error('Token refresh failed');
            }

            return fetch(url, options);
        }

        return response;
    } catch (error) {
        console.error('Fetch error:', error);
        throw error;
    }
}

// Обработчик отправки выражения
async function sendExpression() {
    const inputField = document.getElementById('expression');
    const expression = inputField.value.trim();

    if (!expression) {
        showResultMessage('Введите выражение!', 'error');
        return;
    }

    inputField.value = '';

    try {
        const response = await fetchWithAuth('http://localhost:8080/api/v1/calculate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ expression })
        });

        const data = await response.json();

        if (response.status === 201) {
            showResultMessage(`ID: ${data.id}<br>Результат: <span id="status-${data.id}">В процессе</span>`);
            startPollingResult(data.id);
            startPollingHistory();
        } else {
            showResultMessage(data.error || 'Ошибка вычисления', 'error');
        }
    } catch (error) {
        showResultMessage('Ошибка соединения с сервером', 'error');
    }
}

// Показать сообщение с результатом
function showResultMessage(message, type = 'success') {
    const resultDiv = document.getElementById('result');
    resultDiv.innerHTML = message;
    resultDiv.className = type;
    resultDiv.style.display = 'block';
}

// Запуск опроса результата
function startPollingResult(id) {
    if (resultIntervalId) clearInterval(resultIntervalId);
    resultIntervalId = setInterval(() => pollExpressionResult(id), 1000);
}

// Запуск опроса истории
function startPollingHistory() {
    if (!firstExpressionSent) {
        firstExpressionSent = true;
        if (expressionsIntervalId) clearInterval(expressionsIntervalId);
        expressionsIntervalId = setInterval(fetchExpressions, 5000);
    }
}

// Обработчик Enter
function handleEnterKey(event) {
    if (event.key === 'Enter') {
        sendExpression();
    }
}

// Переключение темы
function toggleTheme() {
    document.body.classList.toggle('dark-theme');
    const theme = document.body.classList.contains('dark-theme') ? 'dark' : 'light';
    localStorage.setItem('theme', theme);
}


document.getElementById('calculateBtn').addEventListener('click', sendExpression);
document.getElementById('expression').addEventListener('keydown', function(event) {
    if (event.key === 'Enter') {
        sendExpression();
    }
});
document.getElementById('themeToggle').addEventListener('click', toggleTheme);

// при загрузке страницы инициализируется список вычислений
document.addEventListener('DOMContentLoaded', () => {
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'dark') {
        document.body.classList.add('dark-theme');
    }
    fetchExpressions();
});

async function fetchExpressions() {
    try {
        const response = await fetchWithAuth(
            'http://localhost:8080/api/v1/expressions',
        {
            method: 'GET',
            headers: {
            'Content-Type': 'application/json',
        },
            credentials: 'include'
        }
        );
        const data = await response.json();
        const historyList = document.getElementById('historyList');
        historyList.innerHTML = "";

        if (response.status === 200) {
            if (data.expressions && Array.isArray(data.expressions) && data.expressions.length > 0) {
                const header = document.createElement('li');
                header.classList.add('history-header');
                header.innerHTML = `
          <span class="history-header-item id-header">ID</span>
          <span class="history-header-item info-header">INFO</span>
        `;
                historyList.appendChild(header);

                data.expressions.sort((a, b) => b.id - a.id).forEach(expr => {
                    const listItem = document.createElement('li');
                    listItem.classList.add('history-item');

                    const idSpan = document.createElement('span');
                    idSpan.classList.add('item-id');
                    idSpan.textContent = `${expr.id}`;

                    const infoDiv = document.createElement('div');
                    infoDiv.classList.add('item-info');

                    const statusDiv = document.createElement('div');
                    statusDiv.classList.add('status');
                    statusDiv.textContent = `status: ${expr.status}`;
                    infoDiv.appendChild(statusDiv);

                    if (expr.result) {
                        const resultDiv = document.createElement('div');
                        resultDiv.classList.add('result');
                        resultDiv.textContent = `result: ${expr.result}`;
                        infoDiv.appendChild(resultDiv);
                    }

                    listItem.appendChild(idSpan);
                    listItem.appendChild(infoDiv);
                    historyList.appendChild(listItem);
                });
            } else {
                const messageLi = document.createElement('li');
                messageLi.classList.add('empty-message');
                messageLi.textContent = 'Пока что тут пусто. Сделай первое вычисление!';
                historyList.appendChild(messageLi);
            }
        } else {
            console.error('Ошибка при получении вычислений');
        }
    } catch (error) {
        console.error('Fetch error during expressions fetch:', error);
        // alert('Ошибка при подключении к серверу');
    }
}

async function pollExpressionResult(id) {
    try {
        const response = await fetchWithAuth(
            `http://localhost:8080/api/v1/expressions/${id}`,
            {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include'
            }
        );
        const data = await response.json();
        console.log('Response from /expressions/{id}:', response);
        console.log('Data from /expressions/{id}:', data);

        if (response.status === 200) {
            const resultSpan = document.getElementById(`status-${id}`);
            if (resultSpan) {
                resultSpan.innerText = data.expression.result || 'В процессе';
                if (data.status === 'done') {
                    clearInterval(resultIntervalId);
                }
            }
        } else {
            console.error('Ошибка при получении результата вычисления');
        }
    } catch (error) {
        console.error('Fetch error during expression result poll:', error);
        // alert('Ошибка при подключении к серверу');
    }
}
