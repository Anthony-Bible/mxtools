describe('DNS Diagnostics', () => {
  it('should allow submitting a domain', () => {
    cy.visit('/dns');
    cy.get('input[type="text"]').type('example.com');
    cy.get('button[type="submit"]').click();
    // Check if loading spinner appears briefly (optional)
    // cy.get('.loading-spinner').should('exist');
    // Check if results area appears (basic check)
    cy.get('pre').should('exist'); 
  });
});
