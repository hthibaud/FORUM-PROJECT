document.addEventListener('DOMContentLoaded', () => {
    const toggleBtn = document.getElementById('theme-toggle');
    const themeIcon = document.getElementById('theme-icon');
    const htmlElement = document.documentElement;
    
    // Fonction pour mettre à jour l'icône
    function updateIcon(theme) {
        if (themeIcon) {
            themeIcon.textContent = theme === 'light' ? '✺' : '⏾';
        }
    }

    // 1. Récupérer la préférence sauvegardée
    // Votre thème par défaut est sombre (les variables CSS sans data-theme)
    const savedTheme = localStorage.getItem('theme') || 'dark';
    
    if (savedTheme === 'light') {
        htmlElement.setAttribute('data-theme', 'light');
    } else {
        htmlElement.removeAttribute('data-theme'); // Laisse le thème par défaut (dark)
    }
    
    updateIcon(savedTheme);

    if (!toggleBtn) return;

    // 2. Gérer le clic sur le bouton
    toggleBtn.addEventListener('click', () => {
        const currentTheme = htmlElement.getAttribute('data-theme');
        let newTheme = 'dark'; // On revient au dark par défaut
        
        // Si c'est le dark par défaut (pas d'attribut) on passe en light
        if (currentTheme !== 'light') {
            newTheme = 'light';
        }
        
        if (newTheme === 'light') {
            htmlElement.setAttribute('data-theme', 'light');
        } else {
            htmlElement.removeAttribute('data-theme');
        }
        
        localStorage.setItem('theme', newTheme);
        updateIcon(newTheme);
    });

    // --- Reply Form Logic ---
    console.log("Setting up reply form listeners...");
    const replyButtons = document.querySelectorAll('.btn-reply');
    console.log(`Found ${replyButtons.length} reply buttons.`);

    replyButtons.forEach(button => {
        button.addEventListener('click', (event) => {
            console.log("Reply button clicked.");
            const commentId = event.currentTarget.dataset.commentId;
            console.log(`Comment ID: ${commentId}`);
            if (!commentId) {
                console.error("Button is missing data-comment-id attribute.");
                return;
            }

            const formId = `reply-form-${commentId}`;
            const form = document.getElementById(formId);
            console.log(`Looking for form with ID: ${formId}`);

            if (form) {
                console.log("Form found. Toggling 'active' class.", form);
                form.classList.toggle('active');
            } else {
                console.error(`Reply form with ID ${formId} not found.`);
            }
        });
    });
});

