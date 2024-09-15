document.addEventListener('DOMContentLoaded', () => {
    const baseUrl = window.location.origin;
    const loginUrl = `${baseUrl}/api/auth/login`;

    fetch(loginUrl, {
        method: 'POST', 
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({}), 
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка при авторизации: ${response.statusText}`);
        }
        return response.json(); 
    })
    .then(data => {
        console.log('Авторизация успешна:', data);
    })
    .catch(error => {
        console.error('Ошибка при авторизации:', error);
    });
});

const dropzone = document.getElementById('dropzone');
const loadingPanel = document.getElementById('loadingPanel');

dropzone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropzone.style.borderColor = '#0056b3';
});

dropzone.addEventListener('dragleave', () => {
    dropzone.style.borderColor = '#007bff';
});

dropzone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropzone.style.borderColor = '#007bff';

    const files = e.dataTransfer.files;
    handleFileUpload(files);
});

dropzone.addEventListener('click', () => {
    const fileInput = document.createElement('input');
    fileInput.type = 'file';
    fileInput.multiple = true;
    fileInput.accept = 'image/*';
    fileInput.addEventListener('change', (e) => {
        const files = e.target.files;
        handleFileUpload(files);
    });
    fileInput.click();
});

function handleFileUpload(files) {
    loadingPanel.style.display = 'block';
    dropzone.style.pointerEvents = 'none';
    dropzone.style.opacity = '0.5'; 

    const fileCount = files.length;
    const baseUrl = window.location.origin;
    const uploadUrl = `${baseUrl}/api/user/getuploadurls`;

    fetch(`${uploadUrl}?count=${fileCount}`, {
        method: 'GET',
        credentials: 'include',
    })
    .then(response => response.json())
    .then(data => {
        console.log('Upload URLs:', data);
        if (data.length !== files.length) {
            throw new Error('Количество файлов не соответствует количеству URL-адресов');
        }

        const uploadPromises = data.map((item, index) => {
            const file = files[index];
            const url = item.Url;

            return fetch(url, {
                method: 'PUT',
                body: file,
                headers: {
                    'Content-Type': file.type,
                },
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error(`Ошибка при загрузке файла ${file.name}: ${response.statusText}`);
                }
                console.log(`Файл ${file.name} успешно загружен.`);
            })
            .catch(error => {
                console.error(`Ошибка при загрузке файла ${file.name}:`, error);
                throw error; 
            });
        });

        Promise.all(uploadPromises)
            .then(() => {
                console.log('Все файлы успешно загружены.');

                return fetch(`${baseUrl}/api/user/completefilesupload`, {
                    method: 'POST',
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(data),
                });
            })
            .catch(error => {
                console.error('Ошибка при загрузке файлов:', error);
            })
            .finally(() => {
                loadingPanel.style.display = 'none';
                dropzone.style.pointerEvents = 'auto';
                dropzone.style.opacity = '1'; 
            });
    })
    .catch(error => {
        console.error('Error fetching upload URLs:', error);

        loadingPanel.style.display = 'none';
        dropzone.style.pointerEvents = 'auto';
        dropzone.style.opacity = '1';
    });
}
