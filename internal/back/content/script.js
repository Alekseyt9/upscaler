

const dropzone = document.getElementById('dropzone');
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
    const fileCount = files.length;
    const baseUrl = window.location.origin;
    const uploadUrl = `${baseUrl}/api/getuploadurls`;

    fetch(`${uploadUrl}?count=${fileCount}`, {
        method: 'GET',
    })
    .then(response => response.json())
    .then(data => {
        console.log('Upload URLs:', data);
                // Проверяем, что количество файлов соответствует количеству URL-адресов
                if (data.length !== files.length) {
                    throw new Error('Количество файлов не соответствует количеству URL-адресов');
                }
        
                // Создаем массив промисов для загрузки файлов
                const uploadPromises = data.map((item, index) => {
                    const file = files[index];
                    const url = item.Url;
        
                    
                    return fetch(url, {
                        method: 'PUT',
                        body: file,
                        headers: {
                            'Content-Type': file.type, 
                            //'x-amz-acl': 'public-read',
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
                    });

                    /*
                    return axios.put(url, file, {
                        headers: {
                            'Content-Type': file.type,
                            // Если нужно добавить дополнительные заголовки, такие как 'x-amz-acl', можно добавить их здесь
                            //'x-amz-acl': 'public-read',
                        },
                        onUploadProgress: (progressEvent) => {
                            const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
                            console.log(`Загрузка ${file.name}: ${percentCompleted}% завершено`);
                        }
                    })
                    .then(response => {
                        console.log(`Файл ${file.name} успешно загружен.`);
                    })
                    .catch(error => {
                        console.error(`Ошибка при загрузке файла ${file.name}:`, error);
                    });*/
                });
        
                // Ждем завершения всех загрузок
                Promise.all(uploadPromises)
                    .then(() => {
                        console.log('Все файлы успешно загружены.');
                    })
                    .catch(error => {
                        console.error('Ошибка при загрузке файлов:', error);
                    });
    })
    .catch(error => {
        console.error('Error fetching upload URLs:', error);
    });
}