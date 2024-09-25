document.addEventListener('DOMContentLoaded', () => {
    const baseUrl = window.location.origin;
    const loginUrl = `${baseUrl}/api/auth/login2`;
    const socket = new WebSocket(loginUrl);

    socket.onopen = () => {
        console.log('WebSocket connection established');
        loadTableData();
    };

    socket.onmessage = (event) => {
        const message = event.data;
        console.log('Message received:', message);

        if (message === 'update') {
            loadTableData();
        }
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
    };

    socket.onclose = () => {
        console.log('WebSocket connection closed');
    };
});

function loadTableData() {
    fetch('/api/user/getstate')
        .then(response => {
            if (!response.ok) {
                throw new Error('Network error');
            }
            return response.json();
        })
        .then(data => {
            createTableRows(data);
        })
        .catch(error => {
            console.error('Error fetching data:', error);
        });
}

function createTableRows(data) {
    const tableBody = document.querySelector('.queue-table tbody');
    tableBody.innerHTML = '';

    if (!data) {
        return;
    }

    data.reverse();

    data.forEach(item => {
        const row = document.createElement('tr');

        const fileNameCell = document.createElement('td');
        fileNameCell.textContent = item.FileName;
        row.appendChild(fileNameCell);

        const linkCell = document.createElement('td');
        if (item.Status == "PROCESSED") {
            const link = document.createElement('a');
            link.href = item.Link;
            link.textContent = 'Download';
            link.setAttribute('download', item.FileName);
            linkCell.appendChild(link);
        } else {
            if (item.Status == "PENDING") {
                linkCell.textContent = item.QueuePosition;
            }
        }
        row.appendChild(linkCell);

        const statusCell = document.createElement('td');
        statusCell.textContent = item.Status;
        switch (item.Status.toLowerCase()) {
            case 'processed':
                statusCell.classList.add('status-processed');
                break;
            case 'pending':
                statusCell.classList.add('status-pending');
                break;
            case 'error':
                statusCell.classList.add('status-error');
                break;
            case 'outdated':
                statusCell.classList.add('status-outdated');
                break;
            default:
                break;
        }
        row.appendChild(statusCell);

        tableBody.appendChild(row);
    });
}

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
            throw new Error('The number of files does not match the number of URLs');
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
                    throw new Error(`Error uploading file ${file.name}: ${response.statusText}`);
                }
                console.log(`File ${file.name} uploaded successfully.`);
            })
            .catch(error => {
                console.error(`Error uploading file ${file.name}:`, error);
                throw error; 
            });
        });

        Promise.all(uploadPromises)
            .then(() => {
                console.log('All files uploaded successfully.');

                fileInfos = [];
                for (var i = 0; i < data.length; i++) {
                    fileInfos.push({
                        Url: data[i].Url, 
                        Key: data[i].Key, 
                        Name: files[i].name,
                    });
                }

                return fetch(`${baseUrl}/api/user/completefilesupload`, {
                    method: 'POST',
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(fileInfos),
                });
            })
            .catch(error => {
                console.error('Error uploading files:', error);
            })
            .finally(() => {
                loadingPanel.style.display = 'none';
                dropzone.style.pointerEvents = 'auto';
                dropzone.style.opacity = '1'; 
                
                loadTableData();
            });
    })
    .catch(error => {
        console.error('Error fetching upload URLs:', error);

        loadingPanel.style.display = 'none';
        dropzone.style.pointerEvents = 'auto';
        dropzone.style.opacity = '1';
    });
}
