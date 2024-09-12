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
    const fileCount = files.length;
    handleFileUpload(fileCount);
});

dropzone.addEventListener('click', () => {
    const fileInput = document.createElement('input');
    fileInput.type = 'file';
    fileInput.multiple = true;
    fileInput.accept = 'image/*';
    fileInput.addEventListener('change', (e) => {
        const files = e.target.files;
        const fileCount = files.length;
        handleFileUpload(fileCount);
    });
    fileInput.click();
});

function handleFileUpload(fileCount) {
    const baseUrl = document.querySelector('meta[name="server-url"]').getAttribute('content');
    const uploadUrl = `${baseUrl}/api/getuploadurls`;

    fetch(`${uploadUrl}?count=${fileCount}`, {
        method: 'GET',
    })
    .then(response => response.json())
    .then(data => {
        console.log('Upload URLs:', data);
    })
    .catch(error => {
        console.error('Error fetching upload URLs:', error);
    });
}