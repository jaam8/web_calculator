let firstExpressionSent = false;
let expressionsIntervalId = null;
let resultIntervalId = null;

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

async function sendExpression() {
    const inputField = document.getElementById('expression');
    const expression = inputField.value.trim();
    if (!expression) {
        alert('Введите выражение!');
        return;
    }
    inputField.value = '';

    try {
        const response = await fetch('http://localhost:8080/api/v1/calculate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ expression })
        });
        const data = await response.json();

        console.log('Response from /calculate:', response);
        console.log('Data from /calculate:', data);

        if (response.status === 201) {
            const resultDiv = document.getElementById('result');
            resultDiv.innerHTML = `
        ID: ${data.id}<br>
        Результат: <span id="status-${data.id}">В процессе</span>
      `;
            resultDiv.style.display = 'block';

            if (resultIntervalId) clearInterval(resultIntervalId);
            resultIntervalId = setInterval(() => pollExpressionResult(data.id), 1000);

            if (!firstExpressionSent) {
                firstExpressionSent = true;
                fetchExpressions();
                if (expressionsIntervalId) clearInterval(expressionsIntervalId);
                expressionsIntervalId = setInterval(fetchExpressions, 1000);
            }
        } else {
            document.getElementById('result').innerText = data.message || 'Ошибка';
            document.getElementById('result').style.display = 'block';
        }
    } catch (error) {
        console.error('Fetch error during expression submission:', error);
        document.getElementById('result').innerText = 'Ошибка при отправке запроса';
        document.getElementById('result').style.display = 'block';
    }
}

async function fetchExpressions() {
    try {
        const response = await fetch('http://localhost:8080/api/v1/expressions');
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
        alert('Ошибка при подключении к серверу');
    }
}

async function pollExpressionResult(id) {
    try {
        const response = await fetch(`http://localhost:8080/api/v1/expressions/${id}`);
        const data = await response.json();
        console.log('Response from /expressions/{id}:', response);
        console.log('Data from /expressions/{id}:', data);

        if (response.status === 200) {
            const resultSpan = document.getElementById(`status-${id}`);
            if (resultSpan) {
                resultSpan.innerText = data.result || 'В процессе';
                if (data.status === 'done') {
                    clearInterval(resultIntervalId);
                }
            }
        } else {
            console.error('Ошибка при получении результата вычисления');
        }
    } catch (error) {
        console.error('Fetch error during expression result poll:', error);
        alert('Ошибка при подключении к серверу');
    }
}

function toggleTheme() {
    document.body.classList.toggle('dark-theme');
    const currentTheme = document.body.classList.contains('dark-theme') ? 'dark' : 'light';
    localStorage.setItem('theme', currentTheme);
}
