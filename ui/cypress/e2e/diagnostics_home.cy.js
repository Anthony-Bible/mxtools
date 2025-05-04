describe('Diagnostics Home', () => {
  it('should display navigation links', () => {
    cy.visit('/');
    cy.contains('DNS Diagnostics');
    cy.contains('Blacklist Check');
    cy.contains('SMTP Diagnostics');
    cy.contains('Email Authentication');
    cy.contains('Network Tools');
  });
});
