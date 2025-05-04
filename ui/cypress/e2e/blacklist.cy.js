describe('Blacklist Check', () => {
  it('should allow submitting an IP', () => {
    cy.visit('/blacklist');
    cy.get('input[type="text"]').type('8.8.8.8');
    cy.get('button[type="submit"]').click();
    cy.get('pre').should('exist');
  });
});
