describe('Email Authentication', () => {
  it('should allow submitting a domain', () => {
    cy.visit('/auth');
    cy.get('input[type="text"]').type('google.com');
    cy.get('button[type="submit"]').click();
    cy.get('pre').should('exist');
  });
});
