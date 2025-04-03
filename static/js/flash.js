// Enhanced flash message handling
function showFlashMessage(message, type = 'error') {
  const flashContainer = document.getElementById('flash-messages');
  const messageElement = document.createElement('div');

  // Set appropriate classes based on message type
  let baseClasses = 'relative mb-4 p-4 rounded-md shadow-md transform transition-all duration-300 ease-in-out translate-y-0 opacity-0';
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
    messageElement.classList.add('translate-y-0', 'opacity-100');
    messageElement.classList.remove('opacity-0');
  }, 10);

  // Auto remove after 5 seconds
  setTimeout(() => {
    removeFlashMessage(messageElement);
  }, 5000);

  return messageElement;
}

function removeFlashMessage(messageElement) {
  messageElement.classList.add('opacity-0', '-translate-y-2');
  setTimeout(() => {
    if (messageElement.parentNode) {
      messageElement.parentNode.removeChild(messageElement);
    }
  }, 300);
}


document.body.addEventListener("showMessage", function (evt) {
  if (evt.detail && evt.detail.level && evt.detail.level === 'flash') {
    showFlashMessage(evt.detail.message, evt.detail.type);
  }
})

// TODO: extract from here
// function handleLoginResponse(event) {
//   const xhr = event.detail.xhr;

//   // Check if the request was successful (status 200-299)
//   if (xhr.status >= 200 && xhr.status < 300) {
//     // For demo purposes, simulate successful login
//     // In a real app, this would be handled by the server response
//     return;
//   }

//   // Handle different error types
//   let errorMessage = 'An unknown error occurred. Please try again.';

//   if (xhr.status === 401) {
//     errorMessage = 'Invalid email or password. Please try again.';
//   } else if (xhr.status === 403) {
//     errorMessage = 'Your account has been locked. Please contact support.';
//   } else if (xhr.status === 429) {
//     errorMessage = 'Too many login attempts. Please try again later.';
//   } else if (xhr.status >= 500) {
//     errorMessage = 'Server error. Please try again later.';
//   }

//   showFlashMessage(errorMessage, 'error');
// }