describe('SMTP Diagnostics', () => {
  it('should allow submitting a domain', () => {
    cy.visit('/smtp');
    cy.get('input[type="text"]').type('gmail.com');
    cy.get('button[type="submit"]').click();
    cy.get('pre').should('exist');
  });
});
