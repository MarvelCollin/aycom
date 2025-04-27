describe('Feed Page', () => {
  beforeEach(() => {
    // Set up authentication state
    localStorage.setItem('access_token', 'fake-access-token');
    localStorage.setItem('refresh_token', 'fake-refresh-token');
    
    // Mock feed API response
    cy.interceptApi('GET', '/feed', {
      success: true,
      posts: [
        {
          id: '1',
          author: { id: 'user1', name: 'John Doe', avatar: 'avatar1.jpg' },
          content: 'This is a test post #1',
          timestamp: new Date().toISOString(),
          likes: 5,
          comments: 2
        },
        {
          id: '2',
          author: { id: 'user2', name: 'Jane Smith', avatar: 'avatar2.jpg' },
          content: 'Another test post with #hashtag',
          timestamp: new Date(Date.now() - 3600000).toISOString(),
          likes: 10,
          comments: 3
        }
      ]
    }).as('feedRequest');
  });
  
  it('should display feed posts correctly', () => {
    // Visit the feed page
    cy.visit('/feed');
    
    // Wait for API call to complete
    cy.wait('@feedRequest');
    
    // Check if posts are displayed
    cy.get('[data-cy=post-item]').should('have.length', 2);
    
    // Check first post content
    cy.get('[data-cy=post-item]').first().within(() => {
      cy.get('[data-cy=author-name]').should('contain', 'John Doe');
      cy.get('[data-cy=post-content]').should('contain', 'This is a test post #1');
      cy.get('[data-cy=like-count]').should('contain', '5');
    });
  });
  
  it('should allow liking a post', () => {
    // Intercept the like API call
    cy.interceptApi('POST', '/posts/*/like', {
      success: true,
      likes: 6
    }).as('likeRequest');
    
    cy.visit('/feed');
    cy.wait('@feedRequest');
    
    // Click the like button on the first post
    cy.get('[data-cy=post-item]').first().within(() => {
      cy.get('[data-cy=like-button]').click();
    });
    
    // Wait for like API call to complete
    cy.wait('@likeRequest');
    
    // Verify like count updated
    cy.get('[data-cy=post-item]').first().within(() => {
      cy.get('[data-cy=like-count]').should('contain', '6');
    });
  });
  
  it('should allow posting a new comment', () => {
    // Intercept the comment API call
    cy.interceptApi('POST', '/posts/*/comment', {
      success: true,
      comment: {
        id: 'comment1',
        author: { id: 'currentUser', name: 'Current User', avatar: 'current-avatar.jpg' },
        content: 'This is a test comment',
        timestamp: new Date().toISOString()
      }
    }).as('commentRequest');
    
    cy.visit('/feed');
    cy.wait('@feedRequest');
    
    // Open comments on the first post
    cy.get('[data-cy=post-item]').first().within(() => {
      cy.get('[data-cy=comment-button]').click();
    });
    
    // Type and submit a comment
    cy.get('[data-cy=comment-input]').type('This is a test comment');
    cy.get('[data-cy=submit-comment]').click();
    
    // Wait for comment API call to complete
    cy.wait('@commentRequest');
    
    // Verify new comment is displayed
    cy.get('[data-cy=comment-item]').should('contain', 'This is a test comment');
  });
  
  it('should handle creating a new post', () => {
    // Intercept the create post API call
    cy.interceptApi('POST', '/posts', {
      success: true,
      post: {
        id: 'new-post-id',
        author: { id: 'currentUser', name: 'Current User', avatar: 'current-avatar.jpg' },
        content: 'My new test post',
        timestamp: new Date().toISOString(),
        likes: 0,
        comments: 0
      }
    }).as('createPostRequest');
    
    cy.visit('/feed');
    cy.wait('@feedRequest');
    
    // Create a new post
    cy.get('[data-cy=new-post-input]').type('My new test post');
    cy.get('[data-cy=post-submit-button]').click();
    
    // Wait for create post API call to complete
    cy.wait('@createPostRequest');
    
    // Verify the new post appears at the top of the feed
    cy.get('[data-cy=post-item]').first().within(() => {
      cy.get('[data-cy=post-content]').should('contain', 'My new test post');
    });
  });
});