<script>
  export let text = '';
  
  function linkify(text) {
    // URL pattern
    const urlRegex = /(https?:\/\/[^\s]+)/g;
    // Twitter-style @username pattern
    const mentionRegex = /(@[a-zA-Z0-9_]+)/g;
    // Twitter-style #hashtag pattern
    const hashtagRegex = /(#[a-zA-Z0-9_]+)/g;

    // Replace URLs with links
    text = text.replace(urlRegex, url => `<a href="${url}" target="_blank" rel="noopener noreferrer">${url}</a>`);
    
    // Replace @mentions with links
    text = text.replace(mentionRegex, mention => {
      const username = mention.substring(1); // Remove the @ symbol
      return `<a href="/user/${username}" class="mention">${mention}</a>`;
    });
    
    // Replace #hashtags with links
    text = text.replace(hashtagRegex, hashtag => {
      const tag = hashtag.substring(1); // Remove the # symbol
      return `<a href="/hashtag/${tag}" class="hashtag">${hashtag}</a>`;
    });
    
    return text;
  }
</script>

<span class="linkified-text">{@html linkify(text)}</span>

<style>
  .linkified-text :global(.hashtag) {
    color: var(--color-primary);
    text-decoration: none;
    font-weight: 500;
    transition: all 0.2s ease;
  }

  .linkified-text :global(.hashtag:hover) {
    text-decoration: underline;
    color: var(--color-primary-dark);
  }

  .linkified-text :global(.mention) {
    color: var(--color-primary);
    text-decoration: none;
    font-weight: 500;
    transition: all 0.2s ease;
  }

  .linkified-text :global(.mention:hover) {
    text-decoration: underline;
    color: var(--color-primary-dark);
  }
</style> 