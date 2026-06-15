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

    // --- Dropdown Menu Logic for Mobile ---
    document.querySelectorAll('.dropdown-toggle').forEach(toggle => {
        toggle.addEventListener('click', event => {
            // Uniquement pour les écrans tactiles/petits
            if (window.innerWidth < 769) {
                event.preventDefault(); // Empêche la navigation
                
                const menu = toggle.nextElementSibling;
                if (menu && menu.classList.contains('dropdown-menu')) {
                    menu.classList.toggle('active');
                    toggle.parentElement.classList.toggle('active');
                }
            }
        });
    });

    window.addEventListener('click', function(e) {
        if (!e.target.matches('.dropdown-toggle')) {
            document.querySelectorAll('.dropdown-menu.active').forEach(menu => {
                menu.classList.remove('active');
                menu.parentElement.classList.remove('active');
            });
        }
    });

    // --- Like / Dislike Logic ---
    document.body.addEventListener('click', (event) => {
        const button = event.target.closest('.like-btn, .dislike-btn');

        if (button) {
            const postID = button.dataset.postId;
            const commentID = button.dataset.commentId;
            const action = parseInt(button.dataset.action, 10);
            
            const isAlreadyActive = button.classList.contains('liked') || button.classList.contains('disliked');
            
            const payload = {
                type: isAlreadyActive ? 0 : action, // Send 0 to remove vote if button is active
            };

            let url = '';
            if (postID) {
                url = '/like/post';
                payload.post_id = parseInt(postID, 10);
            } else if (commentID) {
                url = '/like/comment';
                payload.comment_id = parseInt(commentID, 10);
            } else {
                return; // No ID found
            }

            fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            })
            .then(response => {
                if (!response.ok) {
                    if (response.status === 401) {
                        window.location.href = '/login';
                    }
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                updateLikeUI(button, postID, commentID, data);
            })
            .catch(error => {
                console.error('There has been a problem with your fetch operation:', error);
            });
        }
    });

    function updateLikeUI(button, postID, commentID, data) {
        const container = button.closest('.post-actions') || button.closest('.comment-footer');
        if (!container) return;

        container.querySelector('.likes-count').textContent = data.likes;
        container.querySelector('.dislikes-count').textContent = data.dislikes;

        // Reset both buttons
        const likeBtn = container.querySelector('.like-btn');
        const dislikeBtn = container.querySelector('.dislike-btn');
        likeBtn.classList.remove('liked');
        dislikeBtn.classList.remove('disliked');

        // Apply new state
        if (data.userChoice === 1) {
            likeBtn.classList.add('liked');
        } else if (data.userChoice === -1) {
            dislikeBtn.classList.add('disliked');
        }
    }
});

