// Enhanced flash message handling
export function showFlashMessage(message, type = 'error') {
  const flashContainer = document.getElementById('flash-messages');
  const messageElement = document.createElement('div');

  // Set appropriate classes based on message type
  let baseClasses = 'flash-message relative mb-4 p-4 rounded-md shadow-md';
  let typeClasses = type === 'error'
    ? 'bg-red-50 border border-red-200 text-red-700'
    : 'bg-green-50 border border-green-200 text-green-700';

  messageElement.className = `${baseClasses} ${typeClasses}`;

  // Create message content with icon
  const icon = type === 'error' ? 'fa-circle-exclamation' : 'fa-circle-check';
  messageElement.innerHTML = `
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <i class="fas ${icon} text-lg"></i>
        </div>
        <div class="ml-3">
          <p class="text-sm font-medium">${message}</p>
        </div>
        <div class="ml-auto pl-3">
          <div class="-mx-1.5 -my-1.5">
            <button type="button" class="close-flash inline-flex rounded-md p-1.5 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-red-50 focus:ring-red-600">
              <span class="sr-only">Dismiss</span>
              <i class="fas fa-times"></i>
            </button>
          </div>
        </div>
      </div>
    `;

  // Add close button functionality
  messageElement.querySelector('.close-flash').addEventListener('click', () => {
    removeFlashMessage(messageElement);
  });

  // Add to DOM
  flashContainer.appendChild(messageElement);

  // Animate in
  setTimeout(() => {
    messageElement.classList.add('show');
  }, 10);

  // Auto remove after 5 seconds
  setTimeout(() => {
    removeFlashMessage(messageElement);
  }, 5000);

  return messageElement;
}

export function removeFlashMessage(messageElement) {
  if (!messageElement.classList.contains('show')) return;
  messageElement.classList.remove('show');
  messageElement.classList.add('dismiss');
  setTimeout(() => {
    if (messageElement.parentNode) {
      messageElement.parentNode.removeChild(messageElement);
    }
  }, 150);
}

export function initFlashListeners() {
    document.body.addEventListener("showMessage", function (evt) {
      if (evt.detail && evt.detail.level && evt.detail.level === 'flash') {
        showFlashMessage(evt.detail.message, evt.detail.type);
      }
    })
}
