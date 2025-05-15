// auth.js

document.addEventListener('DOMContentLoaded', () => {
    // Инициализация темы
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'dark') {
        document.body.classList.add('dark-theme');
    }

    // Общая функция обработки ошибок
    const handleAuthError = (errorElementId, error) => {
        const errorElement = document.getElementById(errorElementId);
        errorElement.textContent = error.message || 'Произошла ошибка';
        errorElement.style.display = 'block';
        setTimeout(() => {
            errorElement.style.display = 'none';
        }, 5000);
    };

    // Обработчик для страницы входа
    if (document.getElementById('loginForm')) {
        const loginForm = document.getElementById('loginForm');

        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            const login = document.getElementById('login').value.trim();
            const password = document.getElementById('password').value.trim();
            const errorElement = document.getElementById('loginError');

            if (!login || !password) {
                handleAuthError('loginError', { message: 'Заполните все поля' });
                return;
            }

            try {
                const response = await fetch('http://localhost:8080/api/v1/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ login, password }),
                    credentials: 'include'
                });

                if (response.ok) {
                    window.location.href = '../index.html';
                } else {
                    const errorData = await response.json();
                    handleAuthError('loginError', errorData);
                }
            } catch (error) {
                handleAuthError('loginError', {
                    message: 'Ошибка соединения с сервером'
                });
            }
        });
    }

    // Обработчик для страницы регистрации
    if (document.getElementById('registerForm')) {
        const registerForm = document.getElementById('registerForm');

        registerForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            const login = document.getElementById('regLogin').value.trim();
            const password = document.getElementById('regPassword').value.trim();

            if (!login || !password) {
                handleAuthError('registerError', { message: 'Заполните все поля' });
                return;
            }

            try {
                const response = await fetch('http://localhost:8080/api/v1/register', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ login, password }),
                    credentials: 'include'
                });

                if (response.ok) {
                    window.location.href = 'login.html';
                } else {
                    const errorData = await response.json();
                    handleAuthError('registerError', errorData);
                }
            } catch (error) {
                handleAuthError('registerError', {
                    message: 'Ошибка соединения с сервером'
                });
            }
        });
    }
});